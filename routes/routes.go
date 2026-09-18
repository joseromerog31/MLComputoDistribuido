package routes

import (
	"net/http"

	"worker/controllers"
)

// Registra los endpoints HTTP de cada worker
func SetupRoutes(
	predictionController *controllers.PredictionController,
) *http.ServeMux {

	router := http.NewServeMux()

	// Health Check para verificar que el worker está disponible
	router.HandleFunc(
		"GET /heartbeat",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			w.Write([]byte(`{"status":"ok"}`))
		})

	// Recibe un  batch de registrros y lo manda la controller de ML
	router.HandleFunc(
		"POST /predict-batch",
		predictionController.PredictBatch,
	)

	return router
}
