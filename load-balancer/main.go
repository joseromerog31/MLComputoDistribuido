package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"loadbalancer/balancer"
	"loadbalancer/controllers"
	"loadbalancer/models"
	"loadbalancer/routes"
)

func main() {

	// -----------------------------------------
	// PostgreSQL
	// -----------------------------------------

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

	// -----------------------------------------
	// MODEL
	// -----------------------------------------

	partidoModel :=
		&models.PartidoModel{
			DB: db,
		}

	// -----------------------------------------
	// WORKERS
	// -----------------------------------------

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

	backends :=
		[]*balancer.Backend{}

	for i, target := range urls {

		name :=
			"worker-" +
				string(
					rune('1'+i),
				)

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

	// -----------------------------------------
	// DISTRIBUTOR
	// -----------------------------------------

	loadBalancer :=
		balancer.NewLoadBalancer(
			backends,
		)

	loadBalancer.StartHealthChecks(
		5 * time.Second,
	)

	// -----------------------------------------
	// CONTROLLER
	// -----------------------------------------

	batchController :=
		&controllers.BatchController{
			PartidoModel: partidoModel,

			LoadBalancer: loadBalancer,
		}

	// -----------------------------------------
	// ROUTES
	// -----------------------------------------

	router :=
		routes.SetupRoutes(
			batchController,
		)

	log.Println(
		"Coordinator ejecutándose en :9000",
	)

	err =
		http.ListenAndServe(
			":9000",
			router,
		)

	if err != nil {
		log.Fatal(
			"Error iniciando Coordinator: ",
			err,
		)
	}
}
