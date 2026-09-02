package router

import (
	"github.com/asim9115/containerix/internal/api"
	"github.com/asim9115/containerix/internal/config"
	"github.com/asim9115/containerix/internal/middleware"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/gin-gonic/gin"
)

func NewRouter(repos *repository.Repos, p *pipeline.State, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(
		middleware.MaxBody(cfg.MaxRequestBody),
		middleware.GlobalRateLimit(cfg.GlobalRateLimit, cfg.GlobalRateWindow),
	)
	h := &api.GlobalState{Repos: repos, Pipeline: p, AllowRegistration: cfg.AllowRegistration}

	r.POST("/users",
		middleware.RegistrationRateLimit(cfg.RegistrationRateLimit, cfg.RegistrationRateWindow),
		h.CreateUser,
	)
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
		protected.GET("/containers/:id", h.GetContainer)

		protected.DELETE("/containers/:id", h.DeleteContainer)
		protected.GET("/containers/:id/logs", h.StreamLogs)
		protected.POST("/containers/:id/stop", h.StopContainer)
		protected.POST("/containers/stop-all", h.StopAllContainers)

		// Jobs
		protected.GET("/jobs", h.GetAllJobs)
		protected.GET("/jobs/:id", h.GetJob)


		// User deployments (scoped to authenticated user)
		protected.GET("/deployments", h.ListMyDeployments)
		protected.GET("/deployments/:id", h.GetDeployment)
		protected.DELETE("/deployments/:id", h.DeleteDeployment)

		// Cgroup (admin)
		
	}

	r.GET("/cgroup", h.GetCgroup)
	r.DELETE("/cgroup", h.DeleteCgroup)
	r.GET("/dbports", h.GetPorts)

	r.GET("/health", h.Health)
	r.GET("/ready", h.Ready)
	return r
}
