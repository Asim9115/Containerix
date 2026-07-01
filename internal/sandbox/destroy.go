package sandbox

import (
	"github.com/asim9115/containerix/internal/cgroup"
)

func (s *Sandbox) Destroy() error {
	name := s.Name
	//containers := len(s.Containers)
	path := CgroupRoot
	// if containers != 0 {
	// 	//stop all containers before destroying the cgroup

	// }
	err := cgroup.Destroy(name, path)
	if err != nil {
		return err
	}
	*s = Sandbox{}
	return nil
}