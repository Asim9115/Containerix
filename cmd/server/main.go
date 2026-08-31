package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/asim9115/containerix/internal/config"
	"github.com/asim9115/containerix/internal/database"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/repository/sqllite"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
	"github.com/asim9115/containerix/router"
)

func main() {

    cfg := config.Load()
    // 1. Init database
    if err := database.Init("data/containerix.db"); err != nil {
        log.Fatal(err)
   }
    defer database.Close()

    // 2. Create repositories
    db := database.GetDB()
    repos := &repository.Repos{
       Deployments: sqllite.NewDeploymentRepo(db),
       Jobs:        sqllite.NewJobRepo(db),
       Ports:       sqllite.NewPortsRepo(db),
       User:        sqllite.NewUserRepo(db),
   }

     // 3. Init sandbox (cgroup — stays in-memory, this is OS state)
     if err := state.Init("containerix", cfg.SandboxCPU, cfg.SandboxMemory); err != nil {
         log.Fatal(err)
     }

     // 4. Create pipeline (gets DB access via repos)
     p := pipeline.New(repos)

     //5. Run Reconcile
    data := p.SyncData()
     
    if data != nil {
        state.SB.Sandbox.Allocate(data.CPU, data.Memory)
        
        for port, _ := range data.Ports {
            state.SB.Ports.MarkAsUsed(port)
        }
        // Register running containers into the sandbox map
        for _, c := range data.Containers {
            state.SB.Sandbox.AddContainer(&types.Container{
                ID:     c.ID,
                CPU:    c.CPU,
                Memory: c.Memory,
                Status: "running",
            })
        }
    } else {
        log.Println("[sync] Warning: cgroups or processes not found. Skipping resource pre-allocation.")
    }

    srv := &http.Server{
        Addr: cfg.ListenAddr,
        Handler: router.NewRouter(repos, p),
    }

    go func(){
        log.Printf("server listening on %s", cfg.ListenAddr)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Printf("shutting down...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Forced shutdown:", err)
    }
    log.Println("Server stopped")
}