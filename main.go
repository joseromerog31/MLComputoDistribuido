package main

import (
	"log"
	"net/http"
	"os"

	"worker/controllers"
	"worker/routes"
)

func main() {

	// Obtener nombre de cada instancia del worker
	workerName := os.Getenv("WORKER_NAME")

	if workerName == "" {
		workerName = "worker-local"
	}

	// Controller encargado de cargar el modelo de ML
	predictionController :=
		&controllers.PredictionController{
			WorkerName: workerName,
			ScriptPath: "ml/predict.py",
		}

	// Configurar rutas
	router := routes.SetupRoutes(
		predictionController,
	)

	log.Printf(
		"%s ejecutándose en el puerto 8080",
		workerName,
	)

	// Cada worker escucha el mismo puerto
	// El coordinator maneja a cada contenedor independientemente
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(
			"Error iniciando worker: ",
			err,
		)
	}
}
