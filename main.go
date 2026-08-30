package main

import (
	"log"
	"net/http"

	"crud/controllers"
	"crud/middleware"
	"crud/models"
	"crud/routes"
)

func main() {

	db, err := connectDatabase()

	if err != nil {
		log.Fatal(
			"Error conectando con Postgres: ",
			err,
		)
	}

	defer db.Close()

	log.Println(
		"Conexión con Postgres correcta",
	)

	partidoModel := &models.PartidoModel{
		DB: db,
	}

	partidoController := &controllers.PartidoController{
		PartidoModel: partidoModel,
	}

	router := routes.SetupRoutes(
		partidoController,
	)

	// Middleware
	dispatcher := middleware.NewDispatcher(
		router,
	)

	log.Println(
		"Servidor ejecutándose en http://localhost:8080",
	)

	err = http.ListenAndServe(
		":8080",
		dispatcher,
	)

	if err != nil {
		log.Fatal(
			"Error iniciando el servidor: ",
			err,
		)
	}
}
