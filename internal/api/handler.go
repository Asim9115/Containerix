package api

import (
	"encoding/json"
	"log"
	"net/http"
	"github.com/asim9115/containerix/internal/types"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/state"
	"github.com/gorilla/mux"
)

type githubUrl struct{
	Url string `json:"url"`
}

func CreateDockerImage(w http.ResponseWriter, r *http.Request) {
	var body githubUrl
	//1. Decode the Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Url == "" {
		Error(w, http.StatusBadRequest, "url is required")
		return
	}


	output, err := pipeline.Deploy("first", body.Url)
	
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to deploy")
		
		return
	}
	JSON(w, http.StatusOK, output)
}


func Cgroup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		handleDelete(w, r)
		return

	case http.MethodGet:
		handleGet(w, r)
		return
	case http.MethodPatch:
		handlePatch(w, r)
	default:
		Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	_ = state.SB
	err := state.SB.Sandbox.Destroy()
	if err != nil {
		Error(w, http.StatusConflict, err.Error())
		return
	}
	log.Println("c group deleted")
	JSON(w, http.StatusOK, map[string]string{"Task" :"completed" })
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	// Clean, direct, and safe!
	JSON(w, http.StatusOK, state.SB.Sandbox.GetState())
}

func handlePatch(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		Error(w, http.StatusBadRequest, "No data found")
	}

	
}

func Containers(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, state.SB.Sandbox.GetState().Containers)
}

func StopContainers(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusAccepted, container.StopAll(state.SB.Sandbox.GetState().Containers))
}

func DeleteContainer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	log.Printf("[DeleteContainer] Request received for container id=%q", id)

	// Step 1: look up the container in sandbox state
	containerInfo, exists := state.SB.Sandbox.GetState().Containers[id]
	if !exists {
		log.Printf("[DeleteContainer] FAIL — container id=%q not found in sandbox state", id)
		Error(w, http.StatusBadRequest, "not exists")
		return
	}
	log.Printf("[DeleteContainer] Found container: id=%q name=%q status=%q image=%q",
		containerInfo.ID, containerInfo.Config.Name, containerInfo.Status, containerInfo.Config.Image)

	// Step 2: stop + remove the Docker container
	log.Printf("[DeleteContainer] Calling container.DeleteContainer for docker id=%q", containerInfo.ID)
	err := container.DeleteContainer(containerInfo)
	if err != nil {
		log.Printf("[DeleteContainer] FAIL — container.DeleteContainer error: %v", err)
		Error(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("[DeleteContainer] Docker container id=%q removed successfully", containerInfo.ID)

	// Step 3: release the host port
	if len(containerInfo.Config.Ports) > 0 {
		hostPort := containerInfo.Config.Ports[0].HostPort
		log.Printf("[DeleteContainer] Releasing host port %d", hostPort)
		state.SB.Ports.ReleasePort(hostPort)
		log.Printf("[DeleteContainer] Host port %d released", hostPort)
	} else {
		log.Printf("[DeleteContainer] No ports to release for container id=%q", id)
	}

	// Step 4: release CPU / memory resources from the sandbox
	memory := types.MemoryMap[containerInfo.Config.Tier.Memory]
	log.Printf("[DeleteContainer] Releasing resources — cpu=%.2f memoryKey=%q resolvedMemory=%q",
		containerInfo.Config.Tier.Cpu, containerInfo.Config.Tier.Memory, memory)
	if memory == "" {
		log.Printf("[DeleteContainer] WARN — memory key %q not found in MemoryMap; Release may free 0 bytes",
			containerInfo.Config.Tier.Memory)
	}
	err = state.SB.Sandbox.Release(containerInfo.Config.Tier.Cpu, memory)
	if err != nil {
		log.Printf("[DeleteContainer] FAIL — Sandbox.Release error: %v", err)
		Error(w, http.StatusConflict, err.Error())
		return
	}
	log.Printf("[DeleteContainer] Resources released successfully")

	// Step 5: remove the container record from in-memory sandbox state
	log.Printf("[DeleteContainer] Removing container id=%q from sandbox state", id)
	state.SB.Sandbox.RemoveContainer(id)
	log.Printf("[DeleteContainer] Done — container id=%q deleted successfully", id)

	JSON(w, http.StatusAccepted, "Deleted")
}