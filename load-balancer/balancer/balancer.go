package balancer

import "sync"

type LoadBalancer struct {
	Backends []*Backend
	Current  int

	mu sync.Mutex
}

func NewLoadBalancer(
	backends []*Backend,
) *LoadBalancer {

	return &LoadBalancer{
		Backends: backends,
		Current:  0,
	}
}

func (lb *LoadBalancer) NextBackend() *Backend {

	lb.mu.Lock()
	defer lb.mu.Unlock()

	total := len(lb.Backends)

	if total == 0 {
		return nil
	}

	for i := 0; i < total; i++ {

		index := (lb.Current + i) % total

		backend := lb.Backends[index]

		if backend.IsAlive() {

			lb.Current = (index + 1) % total

			return backend
		}
	}

	return nil
}
