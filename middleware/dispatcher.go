package middleware

import "net/http"

type Job struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Done    chan bool
}

type Dispatcher struct {
	ReadChannel  chan Job
	WriteChannel chan Job
}

func readWorker(readChannel chan Job, router http.Handler) {
	for job := range readChannel {

		router.ServeHTTP(
			job.Writer,
			job.Request,
		)

		job.Done <- true
	}
}

func writeWorker(writeChannel chan Job, router http.Handler) {
	for job := range writeChannel {

		router.ServeHTTP(
			job.Writer,
			job.Request,
		)

		job.Done <- true
	}
}
