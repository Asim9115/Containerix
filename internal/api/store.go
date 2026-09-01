package api

import (
	"sync"

	"github.com/asim9115/containerix/internal/types"
)

// BuildLogRegistry holds in-memory build log buses keyed by job ID.
// Each bus is owned by the deploy pipeline goroutine and removed when it exits.
type BuildLogRegistry struct {
	mu    sync.RWMutex
	buses map[string]*types.LogBus
}

// BuildLogs is the process-global build log registry.
var BuildLogs = &BuildLogRegistry{buses: make(map[string]*types.LogBus)}

func (r *BuildLogRegistry) Set(jobID string, bus *types.LogBus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buses[jobID] = bus
}

func (r *BuildLogRegistry) Get(jobID string) (*types.LogBus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	bus, ok := r.buses[jobID]
	return bus, ok
}

func (r *BuildLogRegistry) Delete(jobID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buses, jobID)
}
