package state

import (
	"fmt"

	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/sandbox"
)

type Server struct {
	Sandbox sandbox.Sandbox
	Ports   *container.Manager
}

var SB Server

// Init initializes the global server state.
// Must be called once at startup before any request is handled.
func Init(name string, cpu float64, memory string) error {
	sb, err := sandbox.Init(name, cpu, memory)
	if err != nil {
		return fmt.Errorf("state: failed to initialize sandbox: %w", err)
	}

	SB.Sandbox = sb
	SB.Ports = container.New()
	return nil
}
