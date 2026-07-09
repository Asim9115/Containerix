package sandbox

import (
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/types"
)

func (s *SandboxManager) UpdateResources(cpu float64, memory string) error {
	err := cgroup.Update(s.Name, cpu, memory, CgroupRoot)
	if err != nil {
		return err
	}
	s.Cpu = cpu
	s.Memory = memory
	return nil
}

func (s *SandboxManager) AddContainer(container *types.Container) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Containers[container.ID] = container
}