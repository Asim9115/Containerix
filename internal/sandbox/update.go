package sandbox

import (
	"github.com/asim9115/containerix/internal/cgroup"
)
func (s *Sandbox) UpdateResources(cpu float64, memory string) error {
	s.Cpu = cpu
	s.Memory = memory

	err := cgroup.Update(s.Name, cpu, memory)
	if err != nil {
		return err
	}
	return nil
}