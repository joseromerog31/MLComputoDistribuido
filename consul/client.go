package consul

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const ServiceName = "ml-worker"

// Client habla con el agente Consul accesible en la red de Docker.
type Client struct {
	address    string
	httpClient *http.Client
}

func NewClient(address string) *Client {
	return &Client{
		address:    strings.TrimRight(address, "/"),
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

type serviceRegistration struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	Address string `json:"Address"`
	Port    int    `json:"Port"`
	Check   struct {
		HTTP                           string `json:"HTTP"`
		Interval                       string `json:"Interval"`
		Timeout                        string `json:"Timeout"`
		DeregisterCriticalServiceAfter string `json:"DeregisterCriticalServiceAfter"`
	} `json:"Check"`
}

func (c *Client) do(ctx context.Context, method, path string, payload any) (*http.Response, error) {
	var reader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.address+path, reader)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		response.Body.Close()
		return nil, fmt.Errorf("Consul respondió %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return response, nil
}

func (c *Client) Register(ctx context.Context, id, host string) error {
	if id == "" || host == "" {
		return fmt.Errorf("ID y host del worker son obligatorios")
	}
	registration := serviceRegistration{
		ID: id, Name: ServiceName, Address: host, Port: 8080,
	}
	registration.Check.HTTP = "http://" + host + ":8080/heartbeat"
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "2s"
	registration.Check.DeregisterCriticalServiceAfter = "1m"

	response, err := c.do(ctx, http.MethodPut,
		"/v1/agent/service/register?replace-existing-checks=true", registration)
	if err != nil {
		return err
	}
	return response.Body.Close()
}

func (c *Client) Deregister(ctx context.Context, id string) error {
	response, err := c.do(ctx, http.MethodPut,
		"/v1/agent/service/deregister/"+url.PathEscape(id), nil)
	if err != nil {
		return err
	}
	return response.Body.Close()
}

// Exists comprueba el registro del agente, no la salud del worker.
func (c *Client) Exists(ctx context.Context, id string) (bool, error) {
	response, err := c.do(ctx, http.MethodGet, "/v1/agent/services", nil)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()
	var services map[string]json.RawMessage
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&services); err != nil {
		return false, err
	}
	_, found := services[id]
	return found, nil
}

// Maintain registra al arrancar y comprueba el registro periódicamente.
// Así se recupera si Consul se reinicia sin datos persistidos.
// Debe ejecutarse en una goroutine y terminar cuando se cancela ctx.
func (c *Client) Maintain(ctx context.Context, id, host string) {
	registered := false
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		if !registered {
			if err := c.Register(ctx, id, host); err != nil {
				if ctx.Err() == nil {
					log.Printf("[CONSUL] registro de %s pendiente: %v", id, err)
				}
			} else {
				registered = true
				log.Printf("[CONSUL] %s registrado como %s", id, ServiceName)
			}
		} else {
			exists, err := c.Exists(ctx, id)
			if err != nil {
				if ctx.Err() == nil {
					log.Printf("[CONSUL] no se pudo consultar el registro: %v", err)
				}
			} else if !exists {
				registered = false
				log.Printf("[CONSUL] %s desapareció del registro; se registrará de nuevo", id)
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}
