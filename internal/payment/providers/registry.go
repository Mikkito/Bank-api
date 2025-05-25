package providers

import (
	"fmt"
	"sync"
)

type Registry struct {
	providers map[string]PaymentProvider
	mu        sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]PaymentProvider),
	}
}

func (r *Registry) Register(p PaymentProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *Registry) Get(name string) (PaymentProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("payment provider not found: %s", name)
	}
	return provider, nil
}

func (r *Registry) GetAll() []PaymentProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]PaymentProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}
