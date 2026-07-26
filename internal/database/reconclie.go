package database

import "github.com/asim9115/containerix/internal/repository"

func Reconcile(repos *repository.Repos) error {
	running, _ := repos.Deployments.ListByStatus("running")
	
}