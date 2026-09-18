package routes

import (
	"net/http"

	"loadbalancer/controllers"
)

func SetupRoutes(
	batchController *controllers.BatchController,
) *http.ServeMux {

	router := http.NewServeMux()

	// Machine Learning distribuido
	router.HandleFunc(
		"POST /predict-batch",
		batchController.PredictBatch,
	)

	// Health del coordinator
	router.HandleFunc(
		"GET /health",
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
				[]byte(
					`{"status":"ok","service":"coordinator"}`,
				),
			)
		},
	)

	return router
}
