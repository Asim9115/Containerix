package api

import (
	"net/http"

	"github.com/asim9115/containerix/internal/database"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

// GlobalState holds all dependencies injected into API handlers.
type GlobalState struct {
	Repos             *repository.Repos
	Pipeline          *pipeline.State
	AllowRegistration bool
}

func (h *GlobalState) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *GlobalState) Ready(c *gin.Context) {
	checks := map[string]string{}
	if err := database.GetDB().Ping(); err != nil {
		checks["database"] = err.Error()
	} else {
		checks["database"] = "ok"
	}
	if anyFailed(checks) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "checks": checks})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

func anyFailed(checks map[string]string) bool {
	for _, val := range checks {
		if val != "ok" {
			return true
		}
	}
	return false
}