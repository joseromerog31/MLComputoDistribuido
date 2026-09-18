package balancer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Estructuras

// Registro que el Coordinator enviará a un worker
type PredictionRecord struct {
	ID                int64   `json:"id"`
	HomeAvgGoalsLast5 float64 `json:"home_avg_goals_last5"`
}

// Formato del batch que recibe cada worker
type BatchRequest struct {
	Records []PredictionRecord `json:"records"`
}

// Predicción individual generada por un worker
type Prediction struct {
	ID         int64   `json:"id"`
	Prediction float64 `json:"prediction"`
}

// Respuesta que devuelve cada worker
type WorkerResponse struct {
	Worker      string       `json:"worker"`
	Processed   int          `json:"processed"`
	Predictions []Prediction `json:"predictions"`
}

// Respuesta final del Coordinator
type BatchResponse struct {
	TotalRecords int              `json:"total_records"`
	WorkersUsed  int              `json:"workers_used"`
	Distribution []WorkerResponse `json:"distribution"`
	Predictions  []Prediction     `json:"predictions"`
}

// Obtener workers disponibles

func (lb *LoadBalancer) GetHealthyBackends() []*Backend {

	healthy := []*Backend{}

	for _, backend := range lb.Backends {

		if backend.IsAlive() {
			healthy = append(
				healthy,
				backend,
			)
		}
	}

	return healthy
}

// Dividir el batch entre los workers

func splitBatch(
	records []PredictionRecord,
	workers int,
) [][]PredictionRecord {

	if len(records) == 0 || workers <= 0 {
		return nil
	}

	// No usamos más workers que registros
	if workers > len(records) {
		workers = len(records)
	}

	chunks := make(
		[][]PredictionRecord,
		workers,
	)

	baseSize := len(records) / workers
	remainder := len(records) % workers

	start := 0

	for i := 0; i < workers; i++ {

		size := baseSize

		// Si la división no es exacta, los primeros workers reciben un registro adicional
		if i < remainder {
			size++
		}

		end := start + size

		chunks[i] = records[start:end]

		start = end
	}

	return chunks
}

// Enviar un chunk a un worker

func sendChunk(
	backend *Backend,
	records []PredictionRecord,
) (WorkerResponse, error) {

	request := BatchRequest{
		Records: records,
	}

	body, err := json.Marshal(request)

	if err != nil {
		return WorkerResponse{}, err
	}

	url := backend.URL.String() +
		"/predict-batch"

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := client.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		return WorkerResponse{},
			fmt.Errorf(
				"no se pudo contactar a %s: %w",
				backend.Name,
				err,
			)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		return WorkerResponse{},
			fmt.Errorf(
				"worker %s respondió %s",
				backend.Name,
				response.Status,
			)
	}

	var workerResponse WorkerResponse

	err = json.NewDecoder(
		response.Body,
	).Decode(&workerResponse)

	if err != nil {
		return WorkerResponse{},
			fmt.Errorf(
				"respuesta inválida de %s: %w",
				backend.Name,
				err,
			)
	}

	return workerResponse, nil
}

// Distribución
func (lb *LoadBalancer) DistributeBatch(
	records []PredictionRecord,
) (BatchResponse, error) {

	// Validar que haya registros.
	if len(records) == 0 {

		return BatchResponse{},
			fmt.Errorf(
				"batch vacío",
			)
	}

	// Obtener únicamente workers sanos
	backends :=
		lb.GetHealthyBackends()

	if len(backends) == 0 {

		return BatchResponse{},
			fmt.Errorf(
				"no hay workers disponibles",
			)
	}

	// Número de workers que realmente usaremos.
	workerCount := len(backends)

	if workerCount > len(records) {
		workerCount = len(records)
	}

	backends =
		backends[:workerCount]

	// Dividir registros entre workers.
	chunks := splitBatch(
		records,
		workerCount,
	)

	// Aquí guardaremos la respuesta
	// de cada worker.
	results := make(
		[]WorkerResponse,
		workerCount,
	)

	// Aquí guardaremos posibles errores.
	errors := make(
		[]error,
		workerCount,
	)

	var wg sync.WaitGroup

	// Concurrenciaaaa

	for i := 0; i < workerCount; i++ {

		wg.Add(1)

		go func(index int) {

			defer wg.Done()

			results[index],
				errors[index] =
				sendChunk(
					backends[index],
					chunks[index],
				)

		}(i)
	}

	// Esperar a que todos los workers terminen
	wg.Wait()

	// Resultados
	allPredictions := []Prediction{}

	for i, workerErr := range errors {

		if workerErr != nil {

			return BatchResponse{},
				fmt.Errorf(
					"error en %s: %w",
					backends[i].Name,
					workerErr,
				)
		}

		allPredictions =
			append(
				allPredictions,
				results[i].Predictions...,
			)
	}

	// Respuesta final
	response := BatchResponse{
		TotalRecords: len(records),
		WorkersUsed:  workerCount,
		Distribution: results,
		Predictions:  allPredictions,
	}

	return response, nil
}
