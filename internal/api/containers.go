package api

import (
	"net/http"

	"github.com/asim9115/containerix/internal/middleware"
	"github.com/gin-gonic/gin"
)

//get user container by id
func (h *GlobalState) GetContainer(c *gin.Context) {
	containerID := c.Param("id")
	container, err := h.Repos.Deployments.GetByContainerId(containerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, container)
}

func (h *GlobalState) DeleteContainer(c *gin.Context) {
	containerID := c.Param("id")
	err := h.Pipeline.DeleteContainer(containerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "container deleted successfully"})
}


//List by user
func (h *GlobalState)GetContainers(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}	
	c.JSON(http.StatusOK, containers)
}

