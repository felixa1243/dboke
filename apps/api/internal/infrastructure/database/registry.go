package database

import (
	"fmt"
	"sync"

	"dboke-api/internal/core/ports"
)

var (
	registryMu       sync.RWMutex
	adapterFactories = make(map[string]func() ports.IDBAdapter)
)

// RegisterAdapter registers a new database adapter factory.
// This should be called in the init() function of the specific adapter package/file.
func RegisterAdapter(name string, factory func() ports.IDBAdapter) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := adapterFactories[name]; exists {
		panic(fmt.Sprintf("Adapter %s already registered", name))
	}
	adapterFactories[name] = factory
}

// GetAdapterFactories returns all registered adapter factories.
func GetAdapterFactories() map[string]func() ports.IDBAdapter {
	registryMu.RLock()
	defer registryMu.RUnlock()
	
	// Return a copy to prevent external mutation
	factoriesCopy := make(map[string]func() ports.IDBAdapter)
	for k, v := range adapterFactories {
		factoriesCopy[k] = v
	}
	return factoriesCopy
}
