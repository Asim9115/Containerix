package api

import (
	"encoding/json"
	"log"
	"net/http"


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
		Status:     types.DeployBuilding,
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
		Status:       types.JobQueued,
		Step:         types.JobQueued,
	}
	if err := h.Repos.Jobs.Create(dbJob); err != nil {
		log.Printf("[handler] failed to create job in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create job"})
		return
	}

	// ── 3. Register build log bus (transient, in-memory only) ────────────────
	logBus := types.NewLogBus()
	BuildLogs.Set(jobId, logBus)

	// ── 4. Launch pipeline in background ─────────────────────────────────────
	go func() {
		defer BuildLogs.Delete(jobId)

		if err := h.Repos.Jobs.UpdateStatus(jobId, types.JobBuilding, "starting pipeline"); err != nil {
			log.Printf("[handler] failed to set job building: %v", err)
		}

		containerID, err := h.Pipeline.Deploy(userID, jobId, logBus, body.Url, tier, body.Env)
		if err != nil {
			if setErr := h.Repos.Jobs.SetFailed(jobId, err.Error()); setErr != nil {
				log.Printf("[handler] failed to mark job failed in DB: %v", setErr)
			}
		} else {
			var hostPort int
			if dep, depErr := h.Repos.Deployments.GetByContainerId(containerID); depErr == nil && dep != nil {
				hostPort = dep.HostPort
			}
			if setErr := h.Repos.Jobs.SetCompleted(jobId, containerID, hostPort); setErr != nil {
				log.Printf("[handler] failed to mark job completed in DB: %v", setErr)
			}
		}
		close(logBus.Ch)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id": jobId,
		"status": types.JobQueued,
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
	userID := c.GetString(middleware.UserIDKey)
	jobs, err := h.Repos.Jobs.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list jobs"})
		return
	}
	c.JSON(http.StatusOK, jobs)
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

