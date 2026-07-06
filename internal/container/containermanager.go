package container

import (
	"github.com/asim9115/containerix/internal/types"
	"github.com/asim9115/containerix/internal/docker"
)
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

func Run(cfg types.Config) (types.Config, error) {
	// Convert raw memory bytes string to docker-friendly format (e.g. "1g")
	newMemory, exists := types.MemoryMap[cfg.Memory]
	if !exists {
		newMemory = "1g"
	}
	cfg.Memory = newMemory

	// cfg.Ports must already be populated by the caller before Run is invoked
	err := docker.RunContainer(cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}