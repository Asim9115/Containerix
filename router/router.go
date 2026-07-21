package router

import (
	"github.com/asim9115/containerix/internal/api"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/build", api.CreateDockerImage)
	r.GET("/cgroup", api.GetCgroup)
	r.DELETE("/cgroup", api.DeleteCgroup)

	r.GET("/containers", api.GetContainers)
	r.GET("/containers/stopall", api.StopContainers)
	r.DELETE("/containers/:id", api.DeleteContainer)
	r.GET("/containers/:id/logs", api.StreamLogs)

	r.GET("/job/:id", api.GetJob)
	return r
}
