package pipeline

import (
	"log"
	"strconv"

	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/types"
)

type SyncContainer struct {
    ID     string
    CPU    float64
    Memory string
}

type Data struct {
	CPU float64
	Memory string
	Ports map[int]string
	Containers []SyncContainer
	}

//Sync the containers with database, performs stopping container if  in host and not in db and updating status if in db and not in host
func (h *State) SyncData() *Data {
	repos := h.Repo
	log.Println("[sync] Running sync data")
	//1. Get the process that are in currently in cgroup.procs
	pids, err := cgroup.GetProcesses()
	if err != nil {
		log.Printf("[sync] processes not found : %v", err)
		return nil
	}
	log.Printf("[sync]process found : %v", pids)
	// 2. Map container IDs currently active on the host
	hostContainersIDMap := make(map[string]bool)
	for _, pid := range pids {
		containerId, err := docker.GetContainerIDFromPID(pid)
		log.Printf("containerId : %v", containerId)
		if err != nil {
			log.Printf("[sync] failed to get containerId of %v", pid)
			continue
		}
		if containerId != "" {
			hostContainersIDMap[containerId] = true
		}
	}

	//3. Get all containers that are in marked as running in database
	dbContainers, err := repos.Deployments.ListByStatus("running")
	if err != nil {
		log.Printf("[sync] error geeting containers from database : %v", err)
		return nil
	}

	//4. map database containers ID
	dbContainersMap := make(map[string]bool)
	for _, dbContainer := range dbContainers {
		if dbContainer.ContainerID != "" {
			dbContainersMap[dbContainer.ContainerID] = true
		}
	}
	//5. reconcile host -> database
	//if its runniong but not in db then stop the container
	for hostContainer := range hostContainersIDMap {
		if !dbContainersMap[hostContainer] {
			log.Printf("[Sync] Stopping orphaned host container: %s", hostContainer)
			if err := container.Stop(hostContainer); err != nil {
				log.Printf("[Sync] Failed to stop container %s: %v", hostContainer, err)
			}
		}
	}

	//6. reconcile database -> host
	//if its marked running in db but not in host container then make the status stopped
	for dbContainer := range dbContainersMap {
		if !hostContainersIDMap[dbContainer] {
			log.Printf("[Sync] Updating out-of-sync DB container to stopped: %s", dbContainer)
			if err := repos.Deployments.UpdateStatusByContainerID(dbContainer, "stopped"); err != nil {
				log.Printf("[Sync] Failed to update deployment status for %s: %v", dbContainer, err)
			}
		}
	}

	SyncedContainers, err := h.Repo.Deployments.ListByStatus("running")
	if err != nil {
		log.Printf("[sync] error getting synced containers from database")
	}

	Data := &Data{
    Ports: make(map[int]string),
}
	 var Memory int
	// calcultate the total cpu and memory to update in sandbox
	for _, container := range SyncedContainers{
		Data.CPU += container.TierCPU
		stringMemory, err := types.MemoryToBytes(container.TierMemory)
		if err != nil {
			log.Printf("[sync]cant convert memory : %v -> %v", container.TierMemory, err)
		}
		newMemory, err :=  strconv.Atoi(stringMemory)
		if err != nil {
			log.Printf("[sync] cant convert to int %v -> %v",newMemory, err)
		}
		Memory += newMemory

		Data.Ports[container.HostPort] = container.ContainerID
		Data.Containers = append(Data.Containers, SyncContainer{
            ID:     container.ContainerID,
            CPU:    container.TierCPU,
            Memory: container.TierMemory,
        }) 
	
	}
	Data.Memory = strconv.Itoa(Memory)
	log.Printf("[sync] sandbox data to sync : %v", Data)
	return Data
}

