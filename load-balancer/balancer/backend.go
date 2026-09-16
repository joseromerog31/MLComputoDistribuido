package balancer

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	Name  string
	URL   *url.URL
	Proxy *httputil.ReverseProxy

	mu    sync.RWMutex
	Alive bool
}

func NewBackend(name string, target string) (*Backend, error) {

	backendURL, err := url.Parse(target)

	if err != nil {
		return nil, err
	}

	backend := &Backend{
		Name:  name,
		URL:   backendURL,
		Alive: true,
	}

	backend.Proxy = httputil.NewSingleHostReverseProxy(
		backendURL,
	)

	return backend, nil
}

func (b *Backend) IsAlive() bool {

	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.Alive
}

func (b *Backend) SetAlive(alive bool) {

	b.mu.Lock()
	defer b.mu.Unlock()

	b.Alive = alive
}
