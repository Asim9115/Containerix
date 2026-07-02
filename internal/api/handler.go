package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"fmt"
	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/detector"
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


	//2. Clone the Repo
	temporaryPath, err := builder.CloneRepository(body.Url)
	if err != nil {
		Error(w, http.StatusBadRequest, "Cannot clone github repository")
		return
	}
	

	//3. scan the code
	// result, err := scanner.ScanFiles(temporaryPath)
	// if err != nil {
	// 	Error(w, http.StatusInternalServerError, "scan failed")
	// 	return
	// }
	// if !result.Clean {
	// 	Error(w, http.StatusBadRequest, result.Error())
	// 	return
	// }


	//4. Detect the language
	detected := detector.Detect(temporaryPath)
	if detected.Language == detector.LangUnknown {
		Error(w, http.StatusBadRequest, "cannot detect language")
		return
	}
	fmt.Println("language detected proceeding to build")
	//5. build docker image
	imageId, err := builder.BuildDockerImage(temporaryPath, detected)
	if err != nil {
		log.Println("docker build error:", err)
		Error(w, http.StatusInternalServerError, "docker build failed")
		return
	}

	//Cleanup
	defer os.RemoveAll(temporaryPath)
	

	//check resouce availability
	err = state.SB.Sandbox.CheckResource(0.5, "524288000")
	if err != nil {
		Error(w, http.StatusUnprocessableEntity, err.Error())
	}

	//runDocker image

	//add process to cgroup

	//update sandbox

	//stream logs
	JSON(w, http.StatusOK, map[string]string{"image_id" : imageId})
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