package sandbox

import (
	"sync"
	"github.com/asim9115/containerix/internal/types"
)

const CgroupRoot = "/sys/fs/cgroup"

type Sandbox interface {
	CanAllocate(cpuNeeded float64, memory string) error
	Allocate(cpu float64, memory string) error
	Release(cpu float64, memory string) error
	UpdateResources(cpu float64, memory string) error
	Destroy() error
	Stats() types.Stats
	Remaining() types.Stats
	GetState() *types.Sandbox
	AddContainer(container *types.Container)
	RemoveContainer(id string)
}

type SandboxManager struct {
	mu sync.Mutex
	*types.Sandbox
}