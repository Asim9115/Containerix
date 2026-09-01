package pipeline

import (
	"fmt"
	"log"

	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
)

func (h *State) DeleteContainer(containerID string) error {
	// 1. Get container configuration to update state and sandbox resources.
	// Try looking up by container_id first, then fallback to deployment id.
	Container, err := h.Repo.Deployments.GetByContainerId(containerID)
	if err != nil {
		return err
	}
	if Container == nil {
		Container, err = h.Repo.Deployments.GetByID(containerID)
		if err != nil {
			return err
		}
	}
	if Container == nil {
		return fmt.Errorf("deployment/container %q not found", containerID)
	}

	// ---------------Stop docker container------------
	if Container.ContainerID != "" {
		err = container.Stop(Container.ContainerID)
		if err != nil {
			log.Printf("[deletecontainer] warning - failed to stop container %s: %v", Container.ContainerID, err)
		}
		// ---------------Delete container from system------
		err = container.DeleteContainer(&types.Container{
			ID:     Container.ContainerID,
			CPU:    Container.TierCPU,
			Memory: Container.TierMemory,
			Status: Container.Status,
		})
		if err != nil {
			log.Printf("[deletecontainer] error deleting container: %v", err)
		}
	}
	//---------------. Free resources----------------------
	err = state.SB.Sandbox.Release(Container.TierCPU, Container.TierMemory)
	if err != nil {
		log.Printf("[deletecontainer] failed to release rsources for container : %v", Container.ContainerID)
		return fmt.Errorf("failed to delete container : %v", err)
	}

	//---------------. Free the port-------------------------
	err = h.Repo.Ports.FreePort(Container.HostPort)
	if err != nil {
		log.Printf("[deletecontainer] failed to free port for container : %v", Container.ContainerID)
		return fmt.Errorf("failed to delete container : %v", err)
	}
	state.SB.Ports.MarkFree(Container.HostPort)

	//--------------Remove container from sandbox map---------
	state.SB.Sandbox.RemoveContainer(containerID)
	if Container.ContainerID != "" {
		state.SB.Sandbox.RemoveContainer(Container.ContainerID)
	}

	//--------------Delete associated jobs and deployment from DB---------
	if err := h.Repo.Jobs.DeleteByDeploymentID(Container.ID); err != nil {
		log.Printf("[deletecontainer] warning - failed to delete jobs for deployment %s: %v", Container.ID, err)
	}

	if err = h.Repo.Deployments.Delete(Container.ID); err != nil {
		log.Printf("[deletecontainer] failed to delete container from db : %v", err)
		return fmt.Errorf("failed to delete container : %v", err)
	}

	return nil

}

func (h *State) StopContainer(container repository.Deployment) error {
	if err := docker.StopContainer(container.ContainerID); err != nil {
		return fmt.Errorf("failed to stop container : %v", err)
	}

	if err := state.SB.Sandbox.Release(container.TierCPU, container.TierMemory); err != nil {
		return err
	}
	state.SB.Sandbox.RemoveContainer(container.ContainerID)
	if container.HostPort > 0 {
		if err := h.Repo.Ports.FreePort(container.HostPort); err != nil {
			log.Printf("[stopcontainer] failed to free port %d in db: %v", container.HostPort, err)
		}
		state.SB.Ports.MarkFree(container.HostPort)
	}

	if err := h.Repo.Deployments.UpdateStatusAndPort(container.ContainerID, types.DeployStopped, 0); err != nil {
		return err
	}
	return nil
	
}

func (h *State) StopAllContainers(userID string) ([]string, error) {
	containers, err := h.Repo.Deployments.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	stopped := make([]string, 0)
	for _, container := range containers {
		if container.Status == types.DeployRunning {
			if err = h.StopContainer(container); err != nil {
				return stopped, err
			}
			stopped = append(stopped, container.ContainerID)
		}
	}
	return stopped, nil
}