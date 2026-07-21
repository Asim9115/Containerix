package api

import (
	"context"
	"fmt"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"time"
)

type BuildRequest struct {
	Url string `json:"url"`
	Tier string `json:"tier"`
	Env map[string]string `json:"env"`
}

var availableTiers = map[string]types.Tier{
	"tier1":types.Tier1,
	"tier2":types.Tier2,
}

func CreateDockerImage(c *gin.Context) {
	var body BuildRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid request body"})
		return
	}
	if body.Url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "url is required"})
		return
	}

	tierName := body.Tier
	if tierName == "" {
		tierName = "tier1"
	}
	tier, ok := availableTiers[tierName]
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tier, use tier1 or tier2"})
		return
	}

	jobId := uuid.New().String()[:8]
	logBus := types.NewLogBus()
	job := &Job{
		ID:        jobId,
		Status:    StatusQueued,
		CreatedAt: time.Now(),
		LogBus:    logBus,
	}
	Jobs.Set(jobId, job)

	go func() {
		Jobs.Update(jobId, func(j *Job) {
			j.Status = StatusBuilding
		})
		containerID, err := pipeline.Deploy(jobId, logBus, body.Url, tier, body.Env)
		if err != nil {
			Jobs.Update(jobId, func(j *Job) {
				j.Status = StatusFailed
				j.Error = err.Error()
			})
		} else {
			containerBus := types.NewLogBus()
			Jobs.Update(jobId, func(j *Job) {
				j.Status = StatusRunning
				j.ContainerID = containerID
				j.ContainerBus = containerBus
			})
			// Stream live container logs in background
			go func() {
				ctx := context.Background()
				_ = docker.StreamContainerLogs(ctx, containerID, containerBus.Ch)
				close(containerBus.Ch)
			}()
		}
		close(logBus.Ch)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobId,
		"status": "queued",
		"logs":   "/containers/" + jobId + "/logs",
	})

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

func StreamLogs(c *gin.Context) {
	id := c.Param("id")
	job, ok := Jobs.Get(id)
	if !ok {
		job, ok = Jobs.GetByContainerID(id)
	}
	if !ok {
		c.JSON(404, gin.H{"error": "job or container not found"})
		return
	}
	// SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}
	// ── Phase A: drain build-time log bus ────────────────────────────────────
	for evt := range job.LogBus.Ch {
		fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
		flusher.Flush()
	}
	// ── Phase B: if deploy failed, close stream ───────────────────────────────
	// Re-read job state after build bus closed
	job, _ = Jobs.Get(id)
	if job.Status == StatusFailed {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", job.Error)
		flusher.Flush()
		return
	}
	// ── Phase C: stream live container logs ───────────────────────────────────
	// ContainerBus may not be assigned yet (tiny race); wait briefly
	var containerBus *types.LogBus
	for i := 0; i < 20; i++ {
	    j, _ := Jobs.Get(id)
	    if j.ContainerBus != nil {
	        containerBus = j.ContainerBus
	        break
	    }
	    time.Sleep(100 * time.Millisecond)
	}
	if containerBus == nil {
	    fmt.Fprintf(c.Writer, "event: done\ndata: container logs unavailable\n\n")
	    flusher.Flush()
	    return
	}
	for evt := range containerBus.Ch {
	    fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
	    flusher.Flush()
	}
	fmt.Fprintf(c.Writer, "event: done\ndata: container stopped\n\n")
	flusher.Flush()
}

func GetJob(c *gin.Context) {
	jobId := c.Param("id")

	job, exists := Jobs.Get(jobId)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"job_id":       job.ID,
		"status":       job.Status,
		"container_id": job.ContainerID,
		"host_port":    job.HostPort,
		"error":        job.Error,
		"created_at":   job.CreatedAt,
		"logs":         "/containers/" + job.ID + "/logs",
	})
}
