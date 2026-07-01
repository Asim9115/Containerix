package state

import "github.com/asim9115/containerix/internal/sandbox"


type Server struct {
	Sandbox *sandbox.Sandbox
}

var SB Server
