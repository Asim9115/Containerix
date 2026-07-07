package api

import (
	"encoding/json"
	"net/http"
	"github.com/asim9115/containerix/internal/pipeline"
	"github.com/asim9115/containerix/internal/state"
)

type githubUrl struct{
	Url string `json:"url"`
}

func CreateDockerImage(w http.ResponseWriter, r *http.Request) {
	var body githubUrl
	//1. Decode the Body
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if body.Url == "" {
		Error(w, http.StatusBadRequest, "url is required")
		return
	}


	output, err := pipeline.Deploy("first", body.Url)
	
	if err != nil {
		Error(w, http.StatusInternalServerError, "failed to deploy")
		
		return
	}
	JSON(w, http.StatusOK, output)
}


func Cgroup(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		handleDelete(w, r)
		return

	case http.MethodGet:
		handleGet(w, r)
		return
	case http.MethodPatch:
		handlePatch(w, r)
	default:
		Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	_ = state.SB
	err := state.SB.Sandbox.Destroy()
	if err != nil {
		Error(w, http.StatusConflict, err.Error())
		return
	}
	JSON(w, http.StatusOK, map[string]string{"Task" :"completed" })
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	JSON(w, http.StatusOK, state.SB.Sandbox.Stats())
}

func handlePatch(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		Error(w, http.StatusBadRequest, "No data found")
	}

	
}