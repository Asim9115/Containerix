package api

import (
	"net/http"

	"github.com/asim9115/containerix/internal/middleware"
	"github.com/gin-gonic/gin"
)

// GET /deployments - Lists deployments to the authenticated user
func (h *GlobalState) ListMyDeployments(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	
	deployments, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get deployments"})
		return
	}
	
	c.JSON(http.StatusOK, deployments)
}
