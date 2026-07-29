package router

import (
	"github.com/asim9115/containerix/internal/api"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

func NewRouter(repos *repository.Repos, p *pipeline.State) *gin.Engine {
	r := gin.Default()
	h := &api.GlobalState{Repos: repos, Pipeline: p}

	r.POST("/build", h.CreateDockerImage)
	r.GET("/cgroup", h.GetCgroup)
	r.DELETE("/cgroup", h.DeleteCgroup)

	r.GET("/containers", h.GetContainers)
	r.GET("/containers/stopall", h.StopContainers)
	r.DELETE("/containers/:id", h.DeleteContainer)
	r.GET("/containers/:id/logs", h.StreamLogs)

	r.GET("/jobs/:id", h.GetJob)
	r.GET("/jobs", h.GetAllJobs)
	return r
}
