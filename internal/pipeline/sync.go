package pipeline

import (
	"log"

	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
)



func SyncData(repos *repository.Repos) {
	//sync all data that is in procs and database, update the status respectively

	//get the process that are in cgroup.procs
	pids, err := cgroup.GetProcesses()
	if err != nil {
		log.Printf("processes not found : %v", err)
		return
	}

	//Get all containers ids from process id
	var containerIDs []string
	for _, pid := range pids {
		containerId, err := docker.GetContainerFromPID(pid)
		if err != nil {
			log.Printf("failed to get containerId of %v", pid)
			continue
		}
		containerIDs = append(containerIDs, containerId)
	}

	//Get all containers that are in database
	dbContainers, err := repos.Deployments.ListByStatus("running")
	if err != nil {
		log.Printf("error geeting containers from database : %v",err)
		return
	}
	var dbContainersID []string
	for _, container := range dbContainers {
		dbContainersID = append(dbContainersID, container.ContainerID)
	}


//Create a map and check the containers in both

	for _, pcontainer := range containerIDs{
		var found = false
		for _, dbcontainer := range dbContainersID {
			if pcontainer == dbcontainer {
				found = true
				break
			}
		}
		if !found {
			container.Stop(pcontainer)
		}
	}
}
