package balancer

type LoadBalancer struct {
	Backends []*Backend
}

func NewLoadBalancer(backends []*Backend) *LoadBalancer {
	return &LoadBalancer{
		Backends: backends,
	}
}
