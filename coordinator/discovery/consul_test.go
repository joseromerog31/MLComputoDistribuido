package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthyWorkers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health/service/ml-worker" || r.URL.Query().Get("passing") != "true" {
			t.Errorf("consulta incorrecta: %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
          {"Service":{"ID":"worker-2","Address":"worker-2","Port":8080}},
          {"Service":{"ID":"worker-1","Address":"worker-1","Port":8080}}
        ]`))
	}))
	defer server.Close()

	backends, err := NewClient(server.URL).HealthyWorkers(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(backends) != 2 {
		t.Fatalf("se esperaban 2 workers; llegaron %d", len(backends))
	}
	if backends[0].Name != "worker-1" || backends[0].URL.String() != "http://worker-1:8080" {
		t.Fatalf("primer worker incorrecto: %#v", backends[0])
	}
}

func TestHealthyWorkersEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()
	backends, err := NewClient(server.URL).HealthyWorkers(context.Background())
	if err != nil || len(backends) != 0 {
		t.Fatalf("resultado inesperado: %v %v", backends, err)
	}
}
