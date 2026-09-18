package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"loadbalancer/balancer"
)

func main() {

	// Obtener URLs de los workers desde Docker Compose
	targets := os.Getenv("BACKEND_URLS")

	// Valores por defecto
	if targets == "" {
		targets =
			"http://worker-1:8080," +
				"http://worker-2:8080," +
				"http://worker-3:8080"
	}

	// Separar las URLs
	urls := strings.Split(
		targets,
		",",
	)

	backends := []*balancer.Backend{}

	// Crear representación de cada worker
	for i, target := range urls {

		name := "worker-" +
			string(rune('1'+i))

		backend, err :=
			balancer.NewBackend(
				name,
				target,
			)

		if err != nil {
			log.Fatal(
				"Error creando backend: ",
				err,
			)
		}

		backends = append(
			backends,
			backend,
		)

		log.Printf(
			"Backend registrado: %s -> %s",
			name,
			target,
		)
	}

	// Crear Load Balancer / Coordinator
	loadBalancer :=
		balancer.NewLoadBalancer(
			backends,
		)

	// Iniciar health checks
	loadBalancer.StartHealthChecks(
		5 * time.Second,
	)

	// Router del coordinador
	router := http.NewServeMux()

	// ------------------------------------------------
	// ML DISTRIBUIDO
	// ------------------------------------------------

	router.HandleFunc(
		"POST /predict-batch",
		loadBalancer.HandleBatch,
	)

	// ------------------------------------------------
	// RESTO DE PETICIONES
	// ------------------------------------------------
	//
	// CRUD, etc. siguen usando el comportamiento
	// normal del Load Balancer.
	//
	router.Handle(
		"/",
		loadBalancer,
	)

	log.Println(
		"Load Balancer / Coordinator ejecutándose en :9000",
	)

	// Iniciar servidor
	err := http.ListenAndServe(
		":9000",
		router,
	)

	if err != nil {
		log.Fatal(
			"Error iniciando Load Balancer: ",
			err,
		)
	}
}
