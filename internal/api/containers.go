package api

import (
	"net/http"


	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/types"
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
	userID := c.GetString(middleware.UserIDKey)
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	containerID := c.Param("id")
	for _, container := range containers {
		if containerID == container.ContainerID{
			err := h.Pipeline.DeleteContainer(containerID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, err)
				return
			}
			c.JSON(http.StatusAccepted,"container deleted successfully")
			return
		}
	}
	c.JSON(http.StatusInternalServerError, err)
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


func (h *GlobalState) StopContainer(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	containerID := c.Param("id")
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	}
	for _, container := range containers{
		if container.ContainerID == containerID {
			if container.Status == types.DeployStopped{
				c.JSON(http.StatusConflict, "container already stopped")
				return
			}
			if container.Status == types.DeployRunning {
				err := h.Pipeline.StopContainer(container)
				if err != nil {
					c.JSON(http.StatusBadRequest, err)
				}
				c.JSON(http.StatusAccepted, "successfully stopped the container")
				return
			}
		}
	} 
	c.JSON(http.StatusNotFound, "container not found")
}

func (h *GlobalState) StopAllContainers(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	err := h.Pipeline.StopAllContainers(userID)
	if err != nil {
		c.JSON(http.StatusConflict, err)
	}
	c.JSON(http.StatusAccepted, "successfully stopped all containers")
}