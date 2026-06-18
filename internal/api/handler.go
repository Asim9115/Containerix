package api

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/scanner"
	"github.com/asim9115/containerix/internal/detector"

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
	defer os.RemoveAll(temporaryPath)

	//3. scan the code
	result, err := scanner.ScanFiles(temporaryPath)
	if err != nil {
		Error(w, http.StatusInternalServerError, "scan failed")
		return
	}
	if !result.Clean {
		Error(w, http.StatusBadRequest, result.Error())
		return
	}


	//4. Detect the language
	detected := detector.DetectLanguage(temporaryPath)
	if detected == detector.LangUnknown {
		Error(w, http.StatusBadRequest, "cannot detect language")
		return
	}

	//5. build docker image
	imageId, err := builder.BuildDockerImage(temporaryPath, detected.Language)
	if err != nil {
    Error(w, http.StatusInternalServerError, "docker build failed")
    return
	}

	JSON(w, http.StatusOK, map[string]string{"image_id" : imageId})
}