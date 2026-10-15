package consul

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterAndDeregister(t *testing.T) {
	operations := make([]string, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		operations = append(operations, r.Method+" "+r.URL.Path)
		if strings.Contains(r.URL.Path, "/register") {
			var svc serviceRegistration
			if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
				t.Error(err)
			}
			if svc.Name != ServiceName || svc.ID != "worker-1" || svc.Address != "worker-1" ||
				svc.Check.HTTP != "http://worker-1:8080/heartbeat" ||
				svc.Check.DeregisterCriticalServiceAfter != "1m" {
				t.Errorf("payload inválido: %+v", svc)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := NewClient(server.URL)
	if err := client.Register(context.Background(), "worker-1", "worker-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.Deregister(context.Background(), "worker-1"); err != nil {
		t.Fatal(err)
	}
	if len(operations) != 2 || operations[0] != "PUT /v1/agent/service/register" ||
		operations[1] != "PUT /v1/agent/service/deregister/worker-1" {
		t.Fatalf("operaciones inesperadas: %v", operations)
	}
}
