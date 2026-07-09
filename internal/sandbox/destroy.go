package sandbox

import (
	"fmt"
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/types"
)



func (s *SandboxManager) Destroy() error {
	if s.Sandbox == nil {
		return fmt.Errorf("sandbox already destroyed or not initialized")
	}

	name := s.Name
	path := CgroupRoot
	err := cgroup.Destroy(name, path, s.Containers)
	if err != nil {
		return err
	}

	// clear fields safely — do NOT zero the whole struct as that nils
	// the embedded *types.Sandbox pointer and causes panics on future calls
	s.Name       = ""
	s.Cpu        = 0
	s.Memory     = ""
	s.UsedCpu    = 0
	s.UsedMemory = "0"
	s.Containers = make(map[string]*types.Container)

	return nil
}