package api

import (
	"fmt"
	"net/http"

	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/types"
	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

func (h *GlobalState) GetContainer(c *gin.Context) {
	containerID := c.Param("id")
	userID := c.GetString(middleware.UserIDKey)
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get container"})
		return
	}

	for _, container := range containers {
		if container.ContainerID == containerID {
			c.JSON(http.StatusOK, container)
			return
		}
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get container"})
}

func (h *GlobalState) DeleteContainer(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list containers"})
		return
	}
	containerID := c.Param("id")
	for _, container := range containers {
		if containerID == container.ContainerID {
			if err := h.Pipeline.DeleteContainer(containerID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Status(http.StatusNoContent)
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
}

func (h *GlobalState) GetContainers(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list containers"})
		return
	}
	c.JSON(http.StatusOK, containers)
}

func (h *GlobalState) StopContainer(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	containerID := c.Param("id")
	containers, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list containers"})
		return
	}
	for _, container := range containers {
		if container.ContainerID != containerID {
			continue
		}
		if container.Status == types.DeployStopped {
			c.JSON(http.StatusConflict, gin.H{"error": "container already stopped"})
			return
		}
		if container.Status != types.DeployRunning {
			c.JSON(http.StatusConflict, gin.H{"error": "container is not running"})
			return
		}
		if err := h.Pipeline.StopContainer(container); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"stopped": containerID})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
}

func (h *GlobalState) StopAllContainers(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	stopped, err := h.Pipeline.StopAllContainers(userID)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stopped": stopped})
}

func (h *GlobalState) StartContainer(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	deploymentID := c.Param("id")
	deployments, err := h.Repos.Deployments.ListByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError ,gin.H{"error": err})
		return
	}
	var data repository.Deployment
	found := false
	for _, deployment := range deployments {
		if deployment.ID == deploymentID {
			data = deployment
			found = true
			break
		}
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error":"deployment not found"})
		return
	}
	if data.ContainerID == "" {
		c.JSON(http.StatusNotFound, gin.H{"error":"container not found"})
		return
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"error inspecting container"})
		return
	}
	defer cli.Close()

	container, exists, err := container.GetContainer(c, cli, data.ContainerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error":"container not found"})
		return
	}

	if container.Container.State.Running {
		c.JSON(http.StatusConflict, gin.H{"error":"container already running"})
		return
	}

	cfg, err := h.Pipeline.StartContainer(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to start container"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"url":     fmt.Sprintf("http://localhost:%d", cfg.Ports[0].HostPort),
	})
}
