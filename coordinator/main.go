package main

import (
	"log"
	"net/http"
	"os"

	"coordinator/controllers"
	"coordinator/discovery"
	"coordinator/models"
	"coordinator/routes"
)

func main() {
	db, err := connectDatabase()
	if err != nil {
		log.Fatal("Error conectando PostgreSQL: ", err)
	}
	defer db.Close()
	log.Println("Coordinator conectado a PostgreSQL")

	partidoModel := &models.PartidoModel{DB: db}
	consulAddress := os.Getenv("CONSUL_HTTP_ADDR")
	if consulAddress == "" {
		consulAddress = "http://localhost:8500"
	}

	batchController := &controllers.BatchController{
		PartidoModel: partidoModel,
		Discovery:    discovery.NewClient(consulAddress),
	}
	router := routes.SetupRoutes(batchController)
	log.Println("Coordinator ejecutándose en :9000; discovery vía Consul")
	if err := http.ListenAndServe(":9000", router); err != nil {
		log.Fatal("Error iniciando Coordinator: ", err)
	}
}
