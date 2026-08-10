package pipeline

import (
	"fmt"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/state"
	"log"
)

func (h *State) DeleteContainer(containerID string) error {
	//1. Get container configuration to update state and sandbox resources
	Container, err := h.Repo.Deployments.GetByContainerId(containerID)
	if err != nil {
		return err
	}
	//---------------Stop docker container------------
	err = container.Stop(Container.ContainerID)
	if err != nil {
		log.Printf("[deletecontainer] failed to stop container : %v", Container.ContainerID)
	}
	//---------------Delete container from system------
	// add delete container later

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

	//--------------Delete the container from DB---------
	if err = h.Repo.Deployments.DeleteByContainerID(containerID); err != nil {
		log.Printf("[deletecontainer] failed to delete container from db : %v", err)
		return fmt.Errorf("failed to delete container : %v", err)
	}

	return nil

}
