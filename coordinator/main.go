package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"coordinator/balancer"
	"coordinator/controllers"
	"coordinator/models"
	"coordinator/routes"
)

func main() {

	// Conectar al coordinator con Postgress
	db, err :=
		connectDatabase()

	if err != nil {
		log.Fatal(
			"Error conectando PostgreSQL: ",
			err,
		)
	}

	defer db.Close()

	log.Println(
		"Coordinator conectado a PostgreSQL",
	)

	// El modelo para partido accesa y toma los registros de la vista de features
	partidoModel :=
		&models.PartidoModel{
			DB: db,
		}

	// Variable de entrono -> direcciones de workers
	targets :=
		os.Getenv("BACKEND_URLS")

	if targets == "" {

		targets =
			"http://worker-1:8080," +
				"http://worker-2:8080," +
				"http://worker-3:8080"
	}

	urls :=
		strings.Split(
			targets,
			",",
		)

	// Registra cada worker como backend para que reciba batches
	backends :=
		[]*balancer.Backend{}

	for i, target := range urls {

		name := fmt.Sprintf("worker-%d", i+1)

		backend, err :=
			balancer.NewBackend(
				name,
				target,
			)

		if err != nil {
			log.Fatal(err)
		}

		backends = append(
			backends,
			backend,
		)

		log.Printf(
			"Worker registrado: %s -> %s",
			name,
			target,
		)
	}

	// Distribuir las predicciones entre los workers que estén arriba
	loadBalancer :=
		balancer.NewLoadBalancer(
			backends,
		)

	// Se corren heath checks periodicas y los workers muertos no figuran para la distribucion
	loadBalancer.StartHealthChecks(
		5 * time.Second,
	)

	// Este controller coordina el flujo de prediccion
	// Postgress -> recupera batches -> workers -> resultados
	batchController :=
		&controllers.BatchController{
			PartidoModel: partidoModel,
			LoadBalancer: loadBalancer,
		}

	// Registra los endpoints del coordinator
	router :=
		routes.SetupRoutes(
			batchController,
		)

	log.Println("Coordinator ejecutándose en :9000")

	// El nginx hace paro para recibir requests y mandarlas
	err = http.ListenAndServe(":9000", router)
	if err != nil {
		log.Fatal("Error iniciando Coordinator: ", err)
	}
}
