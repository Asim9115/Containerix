package router

import (
	"github.com/asim9115/containerix/internal/api"
	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

func NewRouter(repos *repository.Repos, p *pipeline.State) *gin.Engine {
	r := gin.Default()
	h := &api.GlobalState{Repos: repos, Pipeline: p}

	
	r.POST("/users", h.CreateUser) 

	
	protected := r.Group("/")
	protected.Use(middleware.RequestLogger(), middleware.APIKeyAuth(repos))
	{
		// Key management (must be authenticated to rotate your own key)
		protected.POST("/users/api-key", h.RotateAPIKey)
		protected.GET("/users/me", h.GetMe)

		// Build / Deploy
		protected.POST("/build", h.CreateDockerImage)

		// Containers
		protected.GET("/containers", h.GetContainers)
		protected.DELETE("/containers/:id", h.DeleteContainer)
		// protected.GET("/containers/:id/logs", h.StreamLogs)
		protected.GET("/containers/:id", h.GetContainer)
		protected.GET("/containers/stop/:id", h.StopContainer)
		protected.GET("/containers/stopall", h.StopAllContainers)

		// Jobs
		protected.GET("/jobs/:id", h.GetJob)
		protected.GET("/jobs", h.GetAllJobs)

		// User deployments (scoped to authenticated user)
		protected.GET("/deployments", h.ListMyDeployments)

		// Cgroup (admin)
		
		
		
		
	}
	r.GET("/containers/:id/logs", h.StreamLogs)
	r.GET("/cgroup", h.GetCgroup)
	r.DELETE("/cgroup", h.DeleteCgroup)
	r.GET("/dbports", h.GetPorts)
	return r
}
