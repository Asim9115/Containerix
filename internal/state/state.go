package state

import (
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/sandbox"
)


type Server struct {
	Sandbox *sandbox.Sandbox
	Ports *container.Manager
}

var SB Server
