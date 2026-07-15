package api

import (
	"log"
	"net/http"
	"github.com/asim9115/containerix/internal/types"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/state"
    "github.com/gin-gonic/gin"
)

type githubUrl struct{
	Url string `json:"url"`
}

func CreateDockerImage(c *gin.Context) {
    var body githubUrl
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
        return
    }
    if body.Url == "" {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "url is required"})
        return
    }
    output, err := pipeline.Deploy("first", body.Url)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to deploy"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"success": true, "data": output})
}


func GetCgroup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": state.SB.Sandbox.GetState()})
}

func DeleteCgroup(c *gin.Context) {
    if err := state.SB.Sandbox.Destroy(); err != nil {
        c.JSON(http.StatusConflict, gin.H{"success": false, "error": err.Error()})
        return
    }
    log.Println("cgroup deleted")
    c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"Task": "completed"}})
}

func GetContainers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"success": true, "data": state.SB.Sandbox.GetState().Containers})
}

func StopContainers(c *gin.Context) {
    c.JSON(http.StatusAccepted, gin.H{"success": true, "data": container.StopAll(state.SB.Sandbox.GetState().Containers)})
}


func DeleteContainer(c *gin.Context) {
	id := c.Param("id")
	log.Printf("[DeleteContainer] Request received for container id=%q", id)

	// Step 1: look up the container in sandbox state
	containerInfo, exists := state.SB.Sandbox.GetState().Containers[id]
	if !exists {
		log.Printf("[DeleteContainer] FAIL — container id=%q not found in sandbox state", id)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not exists",
		})
		return
	}
	log.Printf("[DeleteContainer] Found container: id=%q name=%q status=%q image=%q",
		containerInfo.ID, containerInfo.Config.Name, containerInfo.Status, containerInfo.Config.Image)

	// Step 2: stop + remove the Docker container
	log.Printf("[DeleteContainer] Calling container.DeleteContainer for docker id=%q", containerInfo.ID)
	err := container.DeleteContainer(containerInfo)
	if err != nil {
		log.Printf("[DeleteContainer] FAIL — container.DeleteContainer error: %v", err)
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
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
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})
		return
	}
	log.Printf("[DeleteContainer] Resources released successfully")

	// Step 5: remove the container record from in-memory sandbox state
	log.Printf("[DeleteContainer] Removing container id=%q from sandbox state", id)
	state.SB.Sandbox.RemoveContainer(id)
	log.Printf("[DeleteContainer] Done — container id=%q deleted successfully", id)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Deleted",
	})
}