package controllers

import (
	"encoding/json"
	"net/http"

	"coordinator/balancer"
	"coordinator/models"
)

// Cordina las request de predicciones entre Postgres y los workers
type BatchController struct {
	PartidoModel *models.PartidoModel
	LoadBalancer *balancer.LoadBalancer
}

// Define cuantos registros deben traerse de Posgress y ser procesados
type PredictBatchRequest struct {
	Limit int `json:"limit"`
}

// Trae registros de Postgres y los convierte en registros de prediccion de workers
// Los distribuye entre los workers que están activos y regresa las predicciones
func (c *BatchController) PredictBatch(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request PredictBatchRequest

	// Hacer el decode del request del cliente
	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"JSON inválido",
			http.StatusBadRequest,
		)
		return
	}

	// 300 de límite por defecto
	if request.Limit == 0 {
		request.Limit = 300
	}

	// Para que no truene
	if request.Limit < 1 {
		http.Error(
			w,
			"limit debe ser mayor a 0",
			http.StatusBadRequest,
		)
		return
	}

	if request.Limit > 5000 {
		http.Error(
			w,
			"limit máximo permitido: 5000",
			http.StatusBadRequest,
		)
		return
	}

	// Obtener registros donde su feature haya sido calculada por la view de Postgres
	partidos, err :=
		c.PartidoModel.GetBatch(
			request.Limit,
		)

	if err != nil {
		http.Error(
			w,
			"Error consultando PostgreSQL: "+
				err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	if len(partidos) == 0 {
		http.Error(
			w,
			"No se encontraron registros para procesar",
			http.StatusNotFound,
		)
		return
	}

	// Convertir datos del Model al formato utilizado por el sistema distribuido
	records := make(
		[]balancer.PredictionRecord,
		0,
		len(partidos),
	)

	for _, partido := range partidos {

		records = append(
			records,
			balancer.PredictionRecord{
				ID:                partido.ID,
				HomeAvgGoalsLast5: partido.HomeAvgGoalsLast5,
			},
		)
	}

	// Cómputo distribuido -> dividir el batch, ejecutar prediccions concurrentemente y agregar sus predicciones
	result, err :=
		c.LoadBalancer.DistributeBatch(
			records,
		)

	if err != nil {
		http.Error(
			w,
			"Error distribuyendo batch: "+
				err.Error(),
			http.StatusBadGateway,
		)
		return
	}

	// View = JSON al cliente
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	err = json.NewEncoder(
		w,
	).Encode(result)

	if err != nil {
		http.Error(
			w,
			"Error generando respuesta",
			http.StatusInternalServerError,
		)
	}
}
