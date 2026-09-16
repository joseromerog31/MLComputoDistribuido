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

	targets := os.Getenv("BACKEND_URLS")

	if targets == "" {

		targets =
			"http://worker-1:8080," +
				"http://worker-2:8080," +
				"http://worker-3:8080"
	}

	urls := strings.Split(
		targets,
		",",
	)

	backends := []*balancer.Backend{}

	for i, target := range urls {

		name := "worker-" +
			string(rune('1'+i))

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
	}

	loadBalancer :=
		balancer.NewLoadBalancer(backends)

	loadBalancer.StartHealthChecks(
		5 * time.Second,
	)

	log.Println(
		"Load Balancer ejecutándose en :9000",
	)

	err := http.ListenAndServe(
		":9000",
		loadBalancer,
	)

	if err != nil {
		log.Fatal(err)
	}
}
