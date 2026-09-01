package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/types"
	"github.com/gin-gonic/gin"
)

// Single SSE connection with two sequential phases:
//  1. Build logs — drain the in-memory build bus while the pipeline runs.
//  2. Container logs — stream docker output scoped to this HTTP request;
//     stops when the client disconnects.
func (h *GlobalState) StreamLogs(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	job, err := h.Repos.Jobs.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query job"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "job not found"})
		return
	}

	flusher, ok := setupSSE(c)
	if !ok {
		return
	}

	// Phase 1: build logs (skip if client connected after build finished).
	if bus, alive := BuildLogs.Get(id); alive {
		for evt := range bus.Ch {
			if err := writeSSE(c.Writer, flusher, evt.Event, evt.Data); err != nil {
				return
			}
		}
	}

	// Phase 2: resolve final job status.
	job, err = h.resolveJob(ctx, id)
	if err != nil {
		return
	}
	if job.Status == types.JobFailed {
		writeSSE(c.Writer, flusher, "error", job.Error)
		return
	}
	if job.Status != types.JobCompleted {
		writeSSE(c.Writer, flusher, "done", "job did not complete")
		return
	}

	// Phase 3: container logs (request-scoped).
	dep, err := h.Repos.Deployments.GetByID(id)
	if err != nil || dep == nil {
		writeSSE(c.Writer, flusher, "error", "deployment not found")
		return
	}
	if dep.ContainerID == "" {
		writeSSE(c.Writer, flusher, "done", "no container")
		return
	}
	if dep.Status != types.DeployRunning {
		writeSSE(c.Writer, flusher, "done", "container not running")
		return
	}

	writeFn := func(evt types.SSEEvent) error {
		return writeSSE(c.Writer, flusher, evt.Event, evt.Data)
	}
	if err := docker.StreamContainerLogs(ctx, dep.ContainerID, writeFn); err != nil && ctx.Err() == nil {
		writeSSE(c.Writer, flusher, "error", err.Error())
		return
	}
	writeSSE(c.Writer, flusher, "done", "stream ended")
}

func setupSSE(c *gin.Context) (http.Flusher, bool) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return nil, false
	}
	return flusher, true
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event, data string) error {
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

// resolveJob returns the current job, waiting if the pipeline is still running.
func (h *GlobalState) resolveJob(ctx context.Context, jobID string) (*repository.Job, error) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		job, err := h.Repos.Jobs.GetByID(jobID)
		if err != nil {
			return nil, err
		}
		if job == nil {
			return nil, fmt.Errorf("job not found")
		}
		if job.Status == types.JobCompleted || job.Status == types.JobFailed {
			return job, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
