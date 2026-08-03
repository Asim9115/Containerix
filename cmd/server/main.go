package main

import (
	"log"
	"net/http"

	"github.com/asim9115/containerix/internal/database"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/repository/sqllite"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/router"
)

func main() {
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
   }

     // 3. Init sandbox (cgroup — stays in-memory, this is OS state)
     if err := state.Init("containerix", 2, "3221225472"); err != nil {
         log.Fatal(err)
     }

     // 4. Create pipeline (gets DB access via repos)
     p := pipeline.New(repos)

     //5. Run Reconcile
    data := p.SyncData()
     
    state.SB.Sandbox.UpdateResources(data.CPU, data.Memory)
    //update host ports

    //replace the container port with actual record or use db 
    
    for port, id := range data.Ports{
        state.SB.Ports.Reserve(id, port, port)
    }
    log.Fatal(http.ListenAndServe(":8080", router.NewRouter(repos, p)))

}