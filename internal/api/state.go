package api

import (
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
)

// GlobalState holds all dependencies injected into API handlers.
type GlobalState struct {
	Repos    *repository.Repos
	Pipeline *pipeline.State
}
