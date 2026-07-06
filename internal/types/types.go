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

var MemoryMap = map[string]string{
	"524288000": "500m", // 500 MB
	"1073741824": "1g",   // 1.0 GB
	"1610612736": "1.5g", // 1.5 GB
	"2147483648": "2g",   // 2.0 GB
	"2684354560": "2.5g", // 2.5 GB
	"3221225472": "3g",   // 3.0 GB
	"3758096384": "3.5g", // 3.5 GB
	"4294967296": "4g",   // 4.0 GB
}