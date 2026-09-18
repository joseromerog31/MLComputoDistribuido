package balancer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Registro individual que se quiere predecir.
type PredictionRecord struct {
	ID                int64   `json:"id"`
	HomeAvgGoalsLast5 float64 `json:"home_avg_goals_last5"`
}

// Batch completo recibido por el coordinador.
type BatchRequest struct {
	Records []PredictionRecord `json:"records"`
}

// Resultado individual de una predicción.
type Prediction struct {
	ID         int64   `json:"id"`
	Prediction float64 `json:"prediction"`
}

// Respuesta que devuelve cada worker.
type WorkerResponse struct {
	Worker      string       `json:"worker"`
	Processed   int          `json:"processed"`
	Predictions []Prediction `json:"predictions"`
}

// Respuesta final que devuelve el coordinador.
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

// Dividir el batch

func splitBatch(
	records []PredictionRecord,
	workers int,
) [][]PredictionRecord {

	if len(records) == 0 || workers <= 0 {
		return nil
	}

	// No tiene sentido usar más workers que registros.
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

		// Los primeros workers reciben un registro adicional si la división no es exacta.
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

	url := backend.URL.String() + "/predict-batch"

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	response, err := client.Post(
		url,
		"application/json",
		bytes.NewReader(body),
	)

	if err != nil {
		return WorkerResponse{}, err
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
		return WorkerResponse{}, err
	}

	return workerResponse, nil
}

// Chamba de load-balancer

func (lb *LoadBalancer) HandleBatch(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request BatchRequest

	err := json.NewDecoder(
		r.Body,
	).Decode(&request)

	if err != nil {

		http.Error(
			w,
			"JSON inválido",
			http.StatusBadRequest,
		)

		return
	}

	// Verificar que haya registros.
	if len(request.Records) == 0 {

		http.Error(
			w,
			"Batch vacío",
			http.StatusBadRequest,
		)

		return
	}

	// Obtener solamente workers activos.
	backends := lb.GetHealthyBackends()

	if len(backends) == 0 {

		http.Error(
			w,
			"No hay workers disponibles",
			http.StatusServiceUnavailable,
		)

		return
	}

	// Número de workers que realmente utilizaremos.
	workerCount := len(backends)

	if workerCount > len(request.Records) {
		workerCount = len(request.Records)
	}

	backends = backends[:workerCount]

	// Dividir registros entre workers.
	chunks := splitBatch(
		request.Records,
		workerCount,
	)

	// Guardar resultados de cada worker.
	results := make(
		[]WorkerResponse,
		workerCount,
	)

	// Guardar posibles errores.
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

			results[index], errors[index] =
				sendChunk(
					backends[index],
					chunks[index],
				)

		}(i)
	}

	// Esperar a que todos los workers terminen.
	wg.Wait()

	// Agregar los resultados

	allPredictions := []Prediction{}

	for i, workerErr := range errors {

		if workerErr != nil {

			http.Error(
				w,
				fmt.Sprintf(
					"Error en %s: %v",
					backends[i].Name,
					workerErr,
				),
				http.StatusBadGateway,
			)

			return
		}

		allPredictions = append(
			allPredictions,
			results[i].Predictions...,
		)
	}

	// Crear respuesta final.
	response := BatchResponse{
		TotalRecords: len(request.Records),
		WorkersUsed:  workerCount,
		Distribution: results,
		Predictions:  allPredictions,
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	err = json.NewEncoder(
		w,
	).Encode(&response)

	if err != nil {
		http.Error(
			w,
			"Error generando respuesta",
			http.StatusInternalServerError,
		)
	}
}
