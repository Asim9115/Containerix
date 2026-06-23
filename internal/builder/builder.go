package builder

import (
	"fmt"
	"os/exec"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"github.com/asim9115/containerix/internal/detector"
)


func CloneRepository(url string) (string, error) {
	//path to store the repository files 
	destPath := filepath.Join(os.TempDir(), uuid.New().String())

	fmt.Println("cloning repository:", url, "→", destPath)

	//executing git clone command
	cmd := exec.Command("git", "clone", url, destPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("git clone failed: %s", string(output))
	}
	return destPath, nil

}

func BuildDockerImage(temporaryPath string, detected detector.DetectResult) (string, error) {
	id := uuid.New()
	tag := "containerix-" + id.String()

	//check if docker file already exists
	if detected.HasDockerfile {
		buildCommand := exec.Command("docker", "build", "-t", tag, temporaryPath)
		output, err := buildCommand.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"error building docker image: %w\n%s",
				err,
				string(output),
			)
		}
		fmt.Printf("Successfully built image %s\n", tag)
		return tag, nil
	}

	//Generate Dockerfile based on detected language
	switch detected.Language {
	case detector.LangNode:
		// Node.js Dockerfile
	case detector.LangPython:
		// Python Dockerfile
	case detector.LangGo:
		// Go Dockerfile
	default:
		return "", fmt.Errorf("unsupported language: %s", detected.Language)
	}

	buildCommand := exec.Command("docker", "build", "-t", tag, temporaryPath)
	output, err := buildCommand.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"error building docker image: %w\n%s",
			err,
			string(output),
		)
	}
	fmt.Printf("Successfully built image %s\n", tag)
	return tag, nil
}