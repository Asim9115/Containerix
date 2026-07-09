package container

import (
	"log"

	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/types"
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

// StopAll stops all containers given a map of containers.
// The caller is responsible for providing the container map (do not read from global state here).
func StopAll(containers map[string]*types.Container) map[string]*types.Container{
	for _, c := range containers {
		err := Stop(c.ID)
		if err != nil {
			log.Printf("failed to stop container %s: %v", c.ID, err)
			continue
		}
		c.Status = "stopped"
	}
	return containers
}