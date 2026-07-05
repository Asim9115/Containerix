package sandbox

import (
	"github.com/asim9115/containerix/internal/cgroup"
)



func (s *SandboxManager)Destroy() error {
	name := s.Name
	path := CgroupRoot
	err := cgroup.Destroy(name, path)
	if err != nil {
		return err
	}
	*s = SandboxManager{}
	return nil
}