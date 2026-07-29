package pipeline

import "github.com/asim9115/containerix/internal/repository"

// State holds all pipeline-level dependencies.
// Every pipeline method gets DB access via h.Repo.
type State struct {
	Repo *repository.Repos
}

// New creates a Pipeline State with the given repos.
func New(repos *repository.Repos) *State {
	return &State{Repo: repos}
}
