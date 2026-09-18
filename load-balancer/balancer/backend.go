package balancer

import (
	"net/url"
	"sync"
)

type Backend struct {
	Name string
	URL  *url.URL

	mu    sync.RWMutex
	Alive bool
}

func NewBackend(name string, target string) (*Backend, error) {
	backendURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &Backend{
		Name:  name,
		URL:   backendURL,
		Alive: true,
	}, nil
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
