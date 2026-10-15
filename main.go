package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"worker/consul"
	"worker/controllers"
	"worker/routes"
)

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	workerName := envOrDefault("WORKER_NAME", "worker-local")
	workerHost := envOrDefault("WORKER_HOST", workerName)
	consulAddress := envOrDefault("CONSUL_HTTP_ADDR", "http://localhost:8500")

	predictionController := &controllers.PredictionController{
		WorkerName: workerName,
		ScriptPath: "ml/predict.py",
	}
	router := routes.SetupRoutes(predictionController)
	server := &http.Server{Addr: ":8080", Handler: router}

	// Reservar el puerto antes de anunciar el worker en Consul.
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Fatalf("Error iniciando %s: %v", workerName, err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	log.Printf("%s ejecutándose en el puerto 8080", workerName)

	consulClient := consul.NewClient(consulAddress)
	maintenanceDone := make(chan struct{})
	go func() {
		defer close(maintenanceDone)
		consulClient.Maintain(ctx, workerName, workerHost)
	}()

	select {
	case <-ctx.Done():
		log.Printf("[WORKER] apagando %s", workerName)
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("[WORKER] error del servidor: %v", err)
		}
	}
	stop() // Detiene el bucle de mantenimiento antes del deregistro.
	<-maintenanceDone

	deregCtx, cancelDereg := context.WithTimeout(context.Background(), 4*time.Second)
	if err := consulClient.Deregister(deregCtx, workerName); err != nil {
		log.Printf("[CONSUL] fallo en deregistro de %s: %v", workerName, err)
	} else {
		log.Printf("[CONSUL] %s eliminado del registro", workerName)
	}
	cancelDereg()

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("[WORKER] cierre no completado: %v", err)
		_ = server.Close()
	}
}
