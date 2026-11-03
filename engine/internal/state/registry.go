package state

import (
	"sync"
)

// GlobalEntityRegistry maintains dynamic live entity pools across generators and DAG streams.
type GlobalEntityRegistry struct {
	mu    sync.RWMutex
	pools map[string][]string
}

var (
	defaultRegistry     *GlobalEntityRegistry
	defaultRegistryOnce sync.Once
)

// GetDefaultRegistry returns the singleton global entity registry.
func GetDefaultRegistry() *GlobalEntityRegistry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = &GlobalEntityRegistry{
			pools: make(map[string][]string),
		}
	})
	return defaultRegistry
}

func NewGlobalEntityRegistry() *GlobalEntityRegistry {
	return &GlobalEntityRegistry{
		pools: make(map[string][]string),
	}
}

// RegisterEntity registers an active entity ID into the designated pool.
func (r *GlobalEntityRegistry) RegisterEntity(poolName, entityID string) {
	if poolName == "" || entityID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pools[poolName] = append(r.pools[poolName], entityID)
}

// GetEntities returns all registered entities in the pool.
func (r *GlobalEntityRegistry) GetEntities(poolName string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := r.pools[poolName]
	res := make([]string, len(entities))
	copy(res, entities)
	return res
}

// SampleEntity selects a live entity from the pool according to deterministic RNG.
func (r *GlobalEntityRegistry) SampleEntity(poolName string, rng *DeterministicRand) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entities := r.pools[poolName]
	if len(entities) == 0 {
		return "", false
	}

	idx := rng.Intn(len(entities))
	return entities[idx], true
}

// PoolSize returns the count of entities registered in the pool.
func (r *GlobalEntityRegistry) PoolSize(poolName string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.pools[poolName])
}

// Clear resets a given pool.
func (r *GlobalEntityRegistry) Clear(poolName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.pools, poolName)
}
