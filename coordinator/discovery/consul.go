// Package discovery obtiene de Consul los workers saludables para cada batch.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"coordinator/balancer"
)

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

type healthServiceEntry struct {
	Node struct {
		Address string `json:"Address"`
	} `json:"Node"`
	Service struct {
		ID      string `json:"ID"`
		Address string `json:"Address"`
		Port    int    `json:"Port"`
	} `json:"Service"`
}

// HealthyWorkers devuelve sólo las instancias passing, consultadas directamente
// al registro. Cada petición de ML crea su propio snapshot estable de workers.
func (c *Client) HealthyWorkers(ctx context.Context) ([]*balancer.Backend, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.address+"/v1/health/service/ml-worker?passing=true", nil)
	if err != nil {
		return nil, err
	}
	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("no se pudo consultar Consul: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		return nil, fmt.Errorf("Consul respondió %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var entries []healthServiceEntry
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&entries); err != nil {
		return nil, fmt.Errorf("respuesta inválida de Consul: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Service.ID < entries[j].Service.ID })

	backends := make([]*balancer.Backend, 0, len(entries))
	for _, entry := range entries {
		host := entry.Service.Address
		if host == "" {
			host = entry.Node.Address
		}
		if host == "" || entry.Service.Port < 1 || entry.Service.Port > 65535 || entry.Service.ID == "" {
			return nil, fmt.Errorf("servicio Consul con dirección, puerto o ID inválido")
		}
		address := "http://" + net.JoinHostPort(host, strconv.Itoa(entry.Service.Port))
		backend, err := balancer.NewBackend(entry.Service.ID, address)
		if err != nil {
			return nil, fmt.Errorf("backend inválido de Consul: %w", err)
		}
		backends = append(backends, backend)
	}
	return backends, nil
}
