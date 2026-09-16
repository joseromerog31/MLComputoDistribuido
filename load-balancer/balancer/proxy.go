package balancer

import (
	"log"
	"net/http"
)

func (lb *LoadBalancer) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	backend := lb.NextBackend()

	if backend == nil {

		http.Error(
			w,
			"No hay workers disponibles",
			http.StatusServiceUnavailable,
		)

		return
	}

	log.Printf(
		"[LOAD BALANCER] %s %s -> %s",
		r.Method,
		r.URL.Path,
		backend.Name,
	)

	w.Header().Set(
		"X-Backend-Worker",
		backend.Name,
	)

	backend.Proxy.ServeHTTP(
		w,
		r,
	)
}
