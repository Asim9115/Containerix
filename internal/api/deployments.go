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

func (h *GlobalState) GetDeployment(c *gin.Context) {
	deploymentID := c.Param("id")
	userID := c.GetString(middleware.UserIDKey)
	deployments, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get deployment"})
		return
	}
	for _, deployment := range deployments {
		if deploymentID == deployment.ID {
			c.JSON(http.StatusOK, deployment)
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get deployment"})
}

func (h *GlobalState) DeleteDeployment(c *gin.Context) {
	deploymentID := c.Param("id")
	userID := c.GetString(middleware.UserIDKey)
	deployments, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get deployment"})
		return
	}
	for _, deployment := range deployments {
		if deploymentID == deployment.ID {
			if err := h.Pipeline.DeleteContainer(deployment.ContainerID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete deployment"})
			}
			c.Status(http.StatusNoContent)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
}
