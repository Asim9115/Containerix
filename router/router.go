package router

import (
	"github.com/asim9115/containerix/internal/api"
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/build", api.CreateDockerImage).Methods("POST")
	r.HandleFunc("/cgroup", api.Cgroup).Methods("GET","DELETE")

	return r
}