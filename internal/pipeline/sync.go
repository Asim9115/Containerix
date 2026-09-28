package pipeline

import (
	"log"
	"strconv"

	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/state"
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

// SyncData reconciles the host cgroup state against the database:
//   - Step 5: containers running on the host but absent from the DB → stop them (orphans).
//   - Step 6: containers marked "running" in the DB but absent from the host → mark them "stopped".
// It returns a Data snapshot of every container that survived both checks so that
// main.go can restore in-memory sandbox resources and port allocations on startup.
func (h *State) SyncData() *Data {
	repos := h.Repo
	log.Println("[sync] Running sync data")
	//1. Get the process that are in currently in cgroup.procs
	pids, err := cgroup.GetProcesses(state.SB.Sandbox.GetState().Name)
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

	// 3. Fetch every deployment the DB considers "running".
	// "active" was the old (wrong) value — the schema and pipeline both use "running".
	dbContainers, err := repos.Deployments.ListByStatus(types.DeployRunning)
	if err != nil {
		log.Printf("[sync] error getting containers from database: %v", err)
		return nil
	}
	allDeployments, err := repos.Deployments.GetAll()
	if err != nil {
		log.Printf("[sync] error getting deployments for port reconciliation: %v", err)
		return nil
	}
	knownContainers := make(map[string]bool, len(allDeployments))
	for _, deployment := range allDeployments {
		if deployment.ContainerID != "" {
			knownContainers[deployment.ContainerID] = true
		}
	}

	//4. map database containers ID
	dbContainersMap := make(map[string]bool)
	dbContainerPorts := make(map[string]int)
	for _, dbContainer := range dbContainers {
		if dbContainer.ContainerID != "" {
			dbContainersMap[dbContainer.ContainerID] = true
			dbContainerPorts[dbContainer.ContainerID] = dbContainer.HostPort
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
			if err := repos.Deployments.UpdateStatusAndPort(dbContainer, types.DeployStopped, dbContainerPorts[dbContainer]); err != nil {
				log.Printf("[Sync] Failed to update deployment status for %s: %v", dbContainer, err)
			}
		}
	}

	// 8. Reconcile port_allocations table.
	// Two scenarios this catches:
	//   a) Missing row  — the server crashed between MarkAsUsed and Ports.Create during a deploy.
	//      The container is alive on the host but has no port_allocations row, so after a restart
	//      the in-memory port manager won't know that port is taken → double allocation on next deploy.
	//   b) Stale row — a deployment was deleted but its port row was left behind.
	existingPorts, portErr := repos.Ports.GetAll()
	if portErr != nil {
		log.Printf("[sync] could not fetch port_allocations for reconcile: %v", portErr)
	} else {
		// Build a set of ports that are already recorded in the DB.
		allocatedInDB := make(map[int]bool, len(existingPorts))
		for _, p := range existingPorts {
			allocatedInDB[p.HostPort] = true
		}

		for _, c := range allDeployments {
			if c.HostPort > 0 && !allocatedInDB[c.HostPort] {
				// Restore reservations for running and stopped containers.
				log.Printf("[sync] re-inserting missing port allocation: host=%d container=%s", c.HostPort, c.ContainerID)
				if err := repos.Ports.Create(&repository.Ports{
					HostPort:      c.HostPort,
					ContainerID:   c.ContainerID,
					ContainerPort: c.ContainerPort,
				}); err != nil {
					log.Printf("[sync] failed to re-insert port %d: %v", c.HostPort, err)
				}
			}
		}

		// Scenario (b): stale rows for containers whose deployment was deleted.
		for _, p := range existingPorts {
			if !knownContainers[p.ContainerID] {
				log.Printf("[sync] removing stale port allocation: host=%d container=%s", p.HostPort, p.ContainerID)
				if err := repos.Ports.DeleteByContainerID(p.ContainerID); err != nil {
					log.Printf("[sync] failed to remove stale port %d: %v", p.HostPort, err)
				}
			}
		}
	}

	// 7. Build the Data snapshot from dbContainers, but only include containers
	// that are actually alive on the host (i.e. survived step 5 & 6 above).
	// We do NOT issue a second ListByStatus("running") query here because step 6
	// may have just mutated some of those rows to "stopped" — a fresh query would
	// race against those writes and could return stale data depending on DB timing.
	data := &Data{
		Ports: make(map[int]string),
	}
	// A stopped Docker container keeps its configured port mapping. Restore
	// those reservations too, so a new deployment cannot claim the same port.
	if existingPorts, err := repos.Ports.GetAll(); err == nil {
		for _, p := range existingPorts {
			if knownContainers[p.ContainerID] {
				data.Ports[p.HostPort] = p.ContainerID
			}
		}
	} else {
		log.Printf("[sync] could not restore port reservations: %v", err)
	}
	var totalMemory int
	for _, c := range dbContainers {
		// Skip containers that step 6 just marked as stopped.
		if !hostContainersIDMap[c.ContainerID] {
			continue
		}
		data.CPU += c.TierCPU

		memBytes, err := types.MemoryToBytes(c.TierMemory)
		if err != nil {
			log.Printf("[sync] can't convert memory %q: %v", c.TierMemory, err)
		} else {
			n, err := strconv.Atoi(memBytes)
			if err != nil {
				log.Printf("[sync] can't parse memory bytes %q: %v", memBytes, err)
			} else {
				totalMemory += n
			}
		}

		data.Ports[c.HostPort] = c.ContainerID
		data.Containers = append(data.Containers, SyncContainer{
			ID:     c.ContainerID,
			CPU:    c.TierCPU,
			Memory: c.TierMemory,
		})
	}
	data.Memory = strconv.Itoa(totalMemory)
	log.Printf("[sync] sandbox data to sync: %+v", data)
	return data
}
