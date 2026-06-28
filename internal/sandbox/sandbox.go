package sandbox

import "github.com/asim9115/containerix/internal/container"

type Sandbox struct {
	Name   string
	Cpu    float64
	Memory string
	UsedCpu float64
	UsedMemory string
	Containers map[string]*container.Container
}

