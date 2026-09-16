package balancer

import (
	"log"
	"net/http"
	"sync"
	"time"
)

func (lb *LoadBalancer) checkBackend(
	backend *Backend,
) {

	client := http.Client{
		Timeout: 2 * time.Second,
	}

	healthURL :=
		backend.URL.String() +
			"/heartbeat"

	response, err := client.Get(
		healthURL,
	)

	alive := err == nil &&
		response.StatusCode == http.StatusOK

	if response != nil {
		response.Body.Close()
	}

	previous := backend.IsAlive()

	backend.SetAlive(alive)

	if previous != alive {

		if alive {
			log.Printf(
				"[HEALTH] %s está disponible",
				backend.Name,
			)
		} else {
			log.Printf(
				"[HEALTH] %s dejó de responder",
				backend.Name,
			)
		}
	}
}

func (lb *LoadBalancer) checkAllBackends() {

	var wg sync.WaitGroup

	for _, backend := range lb.Backends {

		wg.Add(1)

		go func(b *Backend) {

			defer wg.Done()

			lb.checkBackend(b)

		}(backend)
	}

	wg.Wait()
}

func (lb *LoadBalancer) StartHealthChecks(
	interval time.Duration,
) {

	go func() {

		lb.checkAllBackends()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {

			lb.checkAllBackends()
		}
	}()
}
