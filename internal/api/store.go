package api

import (
	"sync"

	"github.com/asim9115/containerix/internal/types"
)

// JobStatus string constants — kept here so handlers can reference them.
type JobStatus string

const (
	StatusQueued   JobStatus = "queued"
	StatusBuilding JobStatus = "building"
	StatusRunning  JobStatus = "running"
	StatusFailed   JobStatus = "failed"
)

type logEntry struct {
	BuildBus     *types.LogBus // emits pipeline log 
	ContainerBus *types.LogBus // emits docker container log 
}

// LogBusRegistry is a minimal in-memory map for SSE channels.
type LogBusRegistry struct {
	mu      sync.RWMutex
	entries map[string]*logEntry
}

// Buses is the process-global registry
var Buses = &LogBusRegistry{entries: make(map[string]*logEntry)}

func (r *LogBusRegistry) Set(id string, build *types.LogBus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries[id] = &logEntry{BuildBus: build}
}

func (r *LogBusRegistry) GetBuild(id string) (*types.LogBus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	if !ok {
		return nil, false
	}
	return e.BuildBus, true
}

func (r *LogBusRegistry) AttachContainerBus(id string, bus *types.LogBus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.entries[id]; ok {
		e.ContainerBus = bus
	}
}

func (r *LogBusRegistry) GetContainerBus(id string) (*types.LogBus, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.entries[id]
	if !ok || e.ContainerBus == nil {
		return nil, false
	}
	return e.ContainerBus, true
}

func (r *LogBusRegistry) Delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, id)
}