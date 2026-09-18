package routes

import (
	"net/http"

	"crud/controllers"
)

func SetupRoutes(
	predictionController *controllers.PredictionController,
) *http.ServeMux {

	router := http.NewServeMux()

	// Health Check
	router.HandleFunc(
		"GET /heartbeat",
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(
				http.StatusOK,
			)

			w.Write(
				[]byte(`{"status":"ok"}`),
			)
		},
	)

	// Machine Learning
	router.HandleFunc(
		"POST /predict-batch",
		predictionController.PredictBatch,
	)

	return router
}
