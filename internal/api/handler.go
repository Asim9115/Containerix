package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BuildRequest struct {
	Url  string            `json:"url"`
	Tier string            `json:"tier"`
	Env  map[string]string `json:"env"`
}

var availableTiers = map[string]types.Tier{
	"tier1": types.Tier1,
	"tier2": types.Tier2,
}

// CreateDockerImage — POST /build
// Immediately persists a Job row with status "queued", then kicks off the
// deploy pipeline in a background goroutine. All status transitions are
// written to the DB so they survive restarts.
func (h *GlobalState) CreateDockerImage(c *gin.Context) {
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
	userID := c.GetString(middleware.UserIDKey)

	// ── 1. Create deployment record first in DB ──────────────────────────────
	var envJSON []byte
	var err error
	if body.Env != nil {
		envJSON, err = json.Marshal(body.Env)
		if err != nil {
			envJSON = []byte("{}")
		}
	} else {
		envJSON = []byte("{}")
	}

	deployment := &repository.Deployment{
		ID:         jobId,
		UserID:     userID,
		RepoURL:    body.Url,
		Status:     "building",
		TierName:   tier.Name,
		TierCPU:    tier.Cpu,
		TierMemory: tier.Memory,
		EnvJSON:    string(envJSON),
	}
	if err := h.Repos.Deployments.Create(deployment); err != nil {
		log.Printf("[handler] failed to create deployment record: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create deployment"})
		return
	}

	// ── 2. Persist job record referencing the deployment (status = queued) ───
	dbJob := &repository.Job{
		ID:           jobId,
		DeploymentID: jobId,
		Status:       string(StatusQueued),
		Step:         "queued",
	}
	if err := h.Repos.Jobs.Create(dbJob); err != nil {
		log.Printf("[handler] failed to create job in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	// ── 3. Register SSE channels (transient, in-memory only) ─────────────────
	logBus := types.NewLogBus()
	Buses.Set(jobId, logBus)

	// ── 3. Launch pipeline in background ─────────────────────────────────────
	go func() {
		defer Buses.Delete(jobId) // clean up SSE entry when goroutine exits

		// Mark as building
		if err := h.Repos.Jobs.UpdateStatus(jobId, string(StatusBuilding), "starting pipeline"); err != nil {
			log.Printf("[handler] failed to set job building: %v", err)
		}

		containerID, err := h.Pipeline.Deploy(userID, jobId, logBus, body.Url, tier, body.Env)
		if err != nil {
			if setErr := h.Repos.Jobs.SetFailed(jobId, err.Error()); setErr != nil {
				log.Printf("[handler] failed to mark job failed in DB: %v", setErr)
			}
		} else {
			// Retrieve the host port that the pipeline recorded on the deployment
			var hostPort int
			if dep, depErr := h.Repos.Deployments.GetByContainerId(containerID); depErr == nil && dep != nil {
				hostPort = dep.HostPort
			}
			if setErr := h.Repos.Jobs.SetCompleted(jobId, containerID, hostPort); setErr != nil {
				log.Printf("[handler] failed to mark job completed in DB: %v", setErr)
			}

			// Attach live container log stream
			containerBus := types.NewLogBus()
			Buses.AttachContainerBus(jobId, containerBus)
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

// GetJob — GET /jobs/:id
func (h *GlobalState) GetJob(c *gin.Context) {
	jobId := c.Param("id")

	job, err := h.Repos.Jobs.GetByID(jobId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query job"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":        job.ID,
		"deployment_id": job.DeploymentID,
		"status":        job.Status,
		"step":          job.Step,
		"error":         job.Error,
		"created_at":    job.CreatedAt,
		"completed_at":  job.CompletedAt,
		"logs":          "/containers/" + job.ID + "/logs",
	})
}

// GetAllJobs — GET /jobs
// Returns all jobs from the DB.
func (h *GlobalState) GetAllJobs(c *gin.Context) {
	jobs, err := h.Repos.Jobs.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list jobs"})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// StreamLogs — GET /containers/:id/logs
// Phase A: drain the build-time SSE channel (in-memory, available only while
//           the goroutine is alive).
// Phase B: if deploy failed, emit error event and close.
// Phase C: stream live container logs from docker.
func (h *GlobalState) StreamLogs(c *gin.Context) {
	id := c.Param("id")

	// Verify job exists in DB
	job, err := h.Repos.Jobs.GetByID(id)
	if err != nil || job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
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

	// ── Phase A: drain build-time log bus (if goroutine still running) ────────
	if buildBus, alive := Buses.GetBuild(id); alive {
		for evt := range buildBus.Ch {
			fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", evt.Event, evt.Data)
			flusher.Flush()
		}
	}

	// ── Phase B: check if job failed (re-read from DB for accuracy) ──────────
	job, _ = h.Repos.Jobs.GetByID(id)
	if job != nil && job.Status == string(StatusFailed) {
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", job.Error)
		flusher.Flush()
		return
	}

	// ── Phase C: stream live container logs ───────────────────────────────────
	// ContainerBus may not be attached yet; poll briefly.
	var containerBus *types.LogBus
	for i := 0; i < 20; i++ {
		if cb, found := Buses.GetContainerBus(id); found {
			containerBus = cb
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

func (h *GlobalState) GetCgroup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": state.SB.Sandbox.GetState()})
}

func (h *GlobalState) DeleteCgroup(c *gin.Context) {
	if err := state.SB.Sandbox.Destroy(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"success": false, "error": err.Error()})
		return
	}
	log.Println("cgroup deleted")
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"Task": "completed"}})
}

func (h *GlobalState) StopContainers(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"success": true, "data": container.StopAll(state.SB.Sandbox.GetState().Containers)})
}