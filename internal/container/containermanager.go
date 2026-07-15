package container

import (
	"log"

	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/types"
)

type ContainerManager interface {
	List(containers map[string]*types.Container) []*types.Container
}

func List(containers map[string]*types.Container) []*types.Container {
	result := make([]*types.Container, 0, len(containers))
	for _, c := range containers {
		result = append(result, c)
	}
	return result
}


//stops a container and the kernel automatically removes the pid from cgroup.procs
func Stop(id string) error {
	err := docker.StopContainer(id)
	if err != nil {
		log.Printf("ContainerManager Error - failed to stop container %s: %v", id, err)
		return err
	}
	return nil
}

//while restarting  a container we need to rewrite the pid of the container in the cgroup
func Start(id string) error {
	err := docker.StartContainer(id)
	if err != nil {
		log.Printf("ContainerManager Error - failed to start container %s: %v", id, err)
		return err
	}
	return nil
}

func Run(cfg types.Config) (types.Config, error) {

	// cfg.Ports must already be populated by the caller before Run is invoked
	err := docker.RunContainer(cfg)
	if err != nil {
		log.Printf("ContainerManager Error - failed to run container %s: %v", cfg.Name, err)
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


func DeleteContainer(container *types.Container) error {
	log.Printf("[container.DeleteContainer] Starting deletion for id=%q status=%q", container.ID, container.Status)

	if container.Status != "stopped" {
		log.Printf("[container.DeleteContainer] Container id=%q is not stopped (status=%q), stopping first", container.ID, container.Status)
		err := Stop(container.ID)
		if err != nil {
			log.Printf("[container.DeleteContainer] FAIL — could not stop container id=%q: %v", container.ID, err)
			return err
		}
		container.Status = "stopped"
		log.Printf("[container.DeleteContainer] Container id=%q stopped successfully", container.ID)
	} else {
		log.Printf("[container.DeleteContainer] Container id=%q already stopped, skipping stop step", container.ID)
	}

	log.Printf("[container.DeleteContainer] Calling docker.DeleteContainer for id=%q", container.ID)
	err := docker.DeleteContainer(container.ID)
	if err != nil {
		log.Printf("[container.DeleteContainer] FAIL — docker.DeleteContainer id=%q: %v", container.ID, err)
		return err
	}
	log.Printf("[container.DeleteContainer] docker.DeleteContainer id=%q succeeded", container.ID)

	return nil
}