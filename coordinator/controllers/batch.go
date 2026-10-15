package controllers

import (
	"encoding/json"
	"net/http"

	"coordinator/balancer"
	"coordinator/discovery"
	"coordinator/models"
)

// BatchController coordina PostgreSQL -> workers -> predicciones.
// Consul es responsable del descubrimiento y health checks.
type BatchController struct {
	PartidoModel *models.PartidoModel
	Discovery    *discovery.Client
}

type PredictBatchRequest struct {
	Limit int `json:"limit"`
}

func (c *BatchController) PredictBatch(w http.ResponseWriter, r *http.Request) {
	var request PredictBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if request.Limit == 0 {
		request.Limit = 300
	}
	if request.Limit < 1 {
		http.Error(w, "limit debe ser mayor a 0", http.StatusBadRequest)
		return
	}
	if request.Limit > 5000 {
		http.Error(w, "limit máximo permitido: 5000", http.StatusBadRequest)
		return
	}

	// Detectar workers ANTES de leer el batch para evitar una consulta
	// a PostgreSQL cuando no queda ninguna instancia disponible.
	backends, err := c.Discovery.HealthyWorkers(r.Context())
	if err != nil {
		http.Error(w, "Error consultando Consul: "+err.Error(), http.StatusBadGateway)
		return
	}
	if len(backends) == 0 {
		http.Error(w, "No hay workers saludables registrados en Consul", http.StatusServiceUnavailable)
		return
	}

	partidos, err := c.PartidoModel.GetBatch(request.Limit)
	if err != nil {
		http.Error(w, "Error consultando PostgreSQL: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if len(partidos) == 0 {
		http.Error(w, "No se encontraron registros para procesar", http.StatusNotFound)
		return
	}

	records := make([]balancer.PredictionRecord, 0, len(partidos))
	for _, partido := range partidos {
		records = append(records, balancer.PredictionRecord{
			ID:                partido.ID,
			HomeAvgGoalsLast5: partido.HomeAvgGoalsLast5,
		})
	}

	// Cada request conserva una lista inmutable de workers.
	// Así no cambiamos la colección mientras DistributeBatch usa goroutines.
	loadBalancer := balancer.NewLoadBalancer(backends)
	result, err := loadBalancer.DistributeBatch(records)
	if err != nil {
		http.Error(w, "Error distribuyendo batch: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		// El encabezado ya pudo haber sido enviado. Solo registrar el error.
		return
	}
}
