package providers

import (
	"fmt"
	"sync"
)

// Registry manages provider instances.
type Registry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register registers a provider with the registry.
func (r *Registry) Register(provider Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[provider.Name()] = provider
}

// Get returns a provider by name.
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return provider, nil
}

// List returns a list of registered provider names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry is the default provider registry.
var DefaultRegistry = NewRegistry()

// Register registers a provider with the default registry.
func Register(provider Provider) {
	DefaultRegistry.Register(provider)
}

// Get returns a provider from the default registry.
func Get(name string) (Provider, error) {
	return DefaultRegistry.Get(name)
}
