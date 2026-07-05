package types

import (
	"sync"
)
// Sandbox is the concrete implementation of the types.Sandbox interface.
type Sandbox struct {
	Name       string                     `json:"name"`
	Cpu        float64                    `json:"cpu"`
	Memory     string                     `json:"memory"`
	UsedCpu    float64                    `json:"usedcpu"`
	UsedMemory string                     `json:"usedmemory"`
	Containers map[string]*Container `json:"containers"`
	mu         sync.Mutex
}



type Stats struct {

	Cpu        float64 `json:"cpu"`
	UsedCpu    float64 `json:"usedcpu"`
	Memory     string  `json:"memory"`
	UsedMemory string  `json:"usedmemory"`
	Containers int     `json:"containers"`
}

type Config struct {
	Name    string
	Image   string
	Cpu     float64
	Memory  string
	Ports   []PortMapping
	Env     map[string]string
	Cmd     []string
	Volumes []VolumeMount
}

type PortMapping struct {
	HostPort      int
	ContainerPort int
}

type VolumeMount struct {
	HostPath      string
	ContainerPath string
}

type Container struct {
	ID     string
	Config Config
	Status string
}
