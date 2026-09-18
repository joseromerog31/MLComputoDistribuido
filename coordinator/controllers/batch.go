package controllers

import (
	"encoding/json"
	"net/http"

	"coordinator/balancer"
	"coordinator/models"
)

type BatchController struct {
	PartidoModel *models.PartidoModel
	LoadBalancer *balancer.LoadBalancer
}

type PredictBatchRequest struct {
	Limit int `json:"limit"`
}

func (c *BatchController) PredictBatch(
	w http.ResponseWriter,
	r *http.Request,
) {

	var request PredictBatchRequest

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

	// 300 de límite por defecto
	if request.Limit == 0 {
		request.Limit = 300
	}

	if request.Limit < 1 {
		http.Error(
			w,
			"limit debe ser mayor a 0",
			http.StatusBadRequest,
		)
		return
	}

	// Para que no truene
	if request.Limit > 5000 {
		http.Error(
			w,
			"limit máximo permitido: 5000",
			http.StatusBadRequest,
		)
		return
	}

	// Obtener datos PostGress
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

	// Cómputo distribuido
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

	// View = JSON
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
