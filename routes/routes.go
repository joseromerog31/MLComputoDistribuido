package routes

import (
	"net/http"

	"crud/controllers"
)

func SetupRoutes(
	partidoController *controllers.PartidoController,
) *http.ServeMux {

	router := http.NewServeMux()

	router.HandleFunc("POST /partidos", partidoController.Create)
	router.HandleFunc("GET /partidos", partidoController.GetAll)
	router.HandleFunc("GET /partidos/{id}", partidoController.GetByID)
	router.HandleFunc("PUT /partidos/{id}", partidoController.Update)
	router.HandleFunc("DELETE /partidos/{id}", partidoController.Delete)

	return router
}
