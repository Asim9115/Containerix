package sandbox

import (
	"github.com/asim9115/containerix/internal/cgroup"
)
func (s *Sandbox) UpdateResources(cpu float64, memory string) error {

	err := cgroup.Update(s.Name, cpu, memory, CgroupRoot)
	if err != nil {
		return err
	}
	s.Cpu = cpu
	s.Memory = memory
	return nil
}