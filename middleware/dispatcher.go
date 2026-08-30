package middleware

import (
	"log"
	"net/http"
)

type Job struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Done    chan bool
}

type Dispatcher struct {
	ReadChannel  chan Job
	WriteChannel chan Job
}

// Worker para operaciones de lectura
func readWorker(readChannel chan Job, router http.Handler) {

	for job := range readChannel {

		log.Printf(
			"[READ WORKER] %s %s",
			job.Request.Method,
			job.Request.URL.Path,
		)

		router.ServeHTTP(
			job.Writer,
			job.Request,
		)

		job.Done <- true
	}
}

// Worker para operaciones de escritura
func writeWorker(writeChannel chan Job, router http.Handler) {

	for job := range writeChannel {

		log.Printf(
			"[WRITE WORKER] %s %s",
			job.Request.Method,
			job.Request.URL.Path,
		)

		router.ServeHTTP(
			job.Writer,
			job.Request,
		)

		job.Done <- true
	}
}

// Crear el dispatcher y levantar los workers
func NewDispatcher(router http.Handler) *Dispatcher {

	dispatcher := &Dispatcher{
		ReadChannel:  make(chan Job),
		WriteChannel: make(chan Job),
	}

	go readWorker(
		dispatcher.ReadChannel,
		router,
	)

	go writeWorker(
		dispatcher.WriteChannel,
		router,
	)

	return dispatcher
}

// Middleware
func (d *Dispatcher) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {

	done := make(chan bool)

	job := Job{
		Writer:  w,
		Request: r,
		Done:    done,
	}

	switch r.Method {

	case http.MethodGet:

		d.ReadChannel <- job

	case http.MethodPost,
		http.MethodPut,
		http.MethodDelete:

		d.WriteChannel <- job

	default:

		http.Error(
			w,
			"Método HTTP no soportado",
			http.StatusMethodNotAllowed,
		)

		return
	}

	<-done
}
