package sandbox

import (
	"sync"

	"github.com/asim9115/containerix/internal/container"
)

const CgroupRoot = "/sys/fs/cgroup"

type Sandbox struct {
	Name       string                          `json:"name"`
	Cpu        float64                         `json:"cpu"`
	Memory     string                          `json:"memory"`
	UsedCpu    float64                         `json:"usedcpu"`
	UsedMemory string                          `json:"usedmemory"`
	Containers map[string]*container.Container `json:"containers"`
	mu 			sync.Mutex
}

type Stats struct {
	Cpu        float64 `json:"cpu"`
	UsedCpu    float64 `json:"usedcpu"`
	Memory     string  `json:"memory"`
	UsedMemory string  `json:"usedmemory"`
	Containers int     `json:"containers"`
}

