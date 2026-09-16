package routes

import (
	"net/http"

	"crud/controllers"
)

func SetupRoutes(
	partidoController *controllers.PartidoController,
) *http.ServeMux {

	router := http.NewServeMux()

	// Health check para que el Load Balancer
	// pueda verificar si el worker está activo
	router.HandleFunc(
		"GET /heartbeat",
		func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			w.Write(
				[]byte(`{"status":"ok"}`),
			)
		},
	)

	// CREATE
	router.HandleFunc(
		"POST /partidos",
		partidoController.Create,
	)

	// READ ALL
	router.HandleFunc(
		"GET /partidos",
		partidoController.GetAll,
	)

	// READ ONE
	router.HandleFunc(
		"GET /partidos/{id}",
		partidoController.GetByID,
	)

	// UPDATE
	router.HandleFunc(
		"PUT /partidos/{id}",
		partidoController.Update,
	)

	// DELETE
	router.HandleFunc(
		"DELETE /partidos/{id}",
		partidoController.Delete,
	)

	return router
}
