package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func main() {
	var url string
	fmt.Println("testing github clone")
	fmt.Println("Enter Url")
	fmt.Scanln(&url)

	cmd := exec.Command("git", "clone", url)
	output , err := cmd.CombinedOutput()
	if err !=  nil {
		fmt.Println(err)
		fmt.Println(string(output))
	}
	fmt.Println("cloning Done")
	 op := buildDockerImage("node","./VirtualStox/nodeSocket" )
	// fmt.Sprintf(err)
	fmt.Println(op)
}

func buildDockerImage(language string, repoPath string) error {
	imageId := uuid.New().String()
	tag := "containerix-" + imageId

	fmt.Println("building Docker image:", tag, "framework", language)

	dockerfilePath := filepath.Join(repoPath, "Dockerfile")

	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		fmt.Println("No Dockerfile found, generating via nixpacks")

		nixCmd := exec.Command("nixpacks", "build", repoPath, "--name", tag)

		output, err := nixCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("nixpacks failed: %v\n%s", err, string(output))
		}

		return nil
	}

	cmd := exec.Command("docker", "build", "-t", tag, repoPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %v\n%s", err, string(output))
	}

	return nil
}