package api

import (
	"sync"
	"time"
	"github.com/asim9115/containerix/internal/types"
)

type JobStatus string

const (
	StatusQueued   JobStatus = "queued"
	StatusBuilding JobStatus = "building"
	StatusRunning  JobStatus = "running"
	StatusFailed   JobStatus = "failed"
)

type Job struct {
	ID          string    `json:"job_id"`
	Status      JobStatus `json:"status"`
	Step        string    `json:"step,omitempty"`
	ContainerID string    `json:"container_id,omitempty"`
	Error       string    `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
    HostPort      int       `json:"host_port,omitempty"`
	//LogBus
	LogBus *types.LogBus `json:"-"`
	ContainerBus  *types.LogBus `json:"-"`
}

type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*Job
}

var Jobs = &JobStore{jobs: make(map[string]*Job)}

func (js *JobStore) Get(id string) (*Job, bool) {
	js.mu.RLock()
	defer js.mu.RUnlock()
	job, ok := js.jobs[id]
	return job, ok
}

func (js *JobStore) Set(id string, job *Job) {
	js.mu.Lock()
	defer js.mu.Unlock()
	js.jobs[id] = job
}

func (js *JobStore) Update(id string, fn func(*Job)) {
	js.mu.Lock()
	defer js.mu.Unlock()
	if j, ok := js.jobs[id]; ok {
		fn(j)
	}
}

func (js *JobStore) GetByContainerID(id string) (*Job, bool) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	for _, job := range js.jobs{
		if job.ContainerID == id {
			return job, true
		}
	}
	return nil, false
}