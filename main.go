package main

import (
	"log"
	"net/http"
	"os"

	"crud/controllers"
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

	workerName := os.Getenv(
		"WORKER_NAME",
	)

	if workerName == "" {
		workerName = "worker-local"
	}

	predictionController :=
		&controllers.PredictionController{
			WorkerName: workerName,
			ScriptPath: "ml/predict.py",
		}

	router := routes.SetupRoutes(
		partidoController,
		predictionController,
	)

	log.Println(
		"Servidor ejecutándose en http://localhost:8080",
	)

	err = http.ListenAndServe(
		":8080",
		router,
	)

	if err != nil {
		log.Fatal(
			"Error iniciando el servidor: ",
			err,
		)
	}
}
