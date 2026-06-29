package sandbox

import "github.com/asim9115/containerix/internal/container"

const cgroupRoot =  "/sys/fs/cgroup" 

type Sandbox struct {
	Name   string
	Cpu    float64
	Memory string
	UsedCpu float64
	UsedMemory string
	Containers map[string]*container.Container
}

type Stats struct {
	Cpu float64
	UsedCpu float64
	Memory string
	UsedMemory string
	Containers int
}