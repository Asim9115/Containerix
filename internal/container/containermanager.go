package container

import "github.com/asim9115/containerix/internal/types"

type ContainerManager interface {
	List() []*types.Container
}

func List(containers map[string]*types.Container) []*types.Container {
	result := make([]*types.Container, 0, len(containers))
	for _, c := range containers {
		result = append(result, c)
	}
	return result
}

func Stop(id string) error {
	err := docker.StopContainer(id)
	if err != nil {
		return err
	}
	return nil
}

func Start(id string) error {
	err := docker.StartContainer(id)
	if err != nil {
		return err
	}
	return nil
}

func Run(cfg types.Config) error {
	err := docker.RunContainer(cfg)
	if err != nil {
		return err
	}
	return nil
}