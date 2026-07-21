package builder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"net/url"
	"github.com/google/uuid"
)


func CloneRepository(repoUrl string) (string, error) {

	//path to store the repository files 
	destPath := filepath.Join("tmp", uuid.New().String())

    if err := os.MkdirAll("tmp", 0755); err != nil {
		log.Printf("Builder Error - Failed to create tmp directory: %v", err)
        return "", err
    }
	fmt.Println("cloning repository:", repoUrl, "→", destPath)

	//executing git clone command
	cmd := exec.Command("git", "clone", repoUrl, destPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("Builder Error - Git clone failed for %s: %v. Output: %s", repoUrl, err, string(output))
		return "", fmt.Errorf("git clone failed: %s", string(output))
	}
	fmt.Println("Clone success")
	return destPath, nil

}

func BuildDockerImage(temporaryPath string) (string, error) {
	id := uuid.New()
	tag := "ctx-" + id.String()[:8]
	
	//if dockerfile exists then go with dockerfile
	dockerFilePath := filepath.Join(temporaryPath, "Dockerfile")
	DockerFilePath := filepath.Join(temporaryPath, "dockerfile")
	var validDockerFilePath string

	if fileExists(dockerFilePath) {
		validDockerFilePath = dockerFilePath
	} else if fileExists(DockerFilePath) {
		validDockerFilePath = DockerFilePath
	}
	log.Print("detecting dockefile")
	if validDockerFilePath != "" {
		log.Print("Dockerfile exists")
		log.Printf("Builder : running docker build for path=%s tag=%s", temporaryPath, tag)
		cmd := exec.Command("docker", "build", "-f", validDockerFilePath, "-t", tag, temporaryPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("Builder Error - docker build failed for %s: %v\n", tag, err)
			return "", fmt.Errorf("docker build failed: %w\n", err)
		}

		return tag, nil
	}

	log.Printf("Builder: running nixpacks build for path=%s tag=%s", temporaryPath, tag)

	cmd := exec.Command("nixpacks", "build", temporaryPath, "--name", tag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("Builder Error - nixpacks build failed for %s: %v\n", tag, err)
		return "", fmt.Errorf("nixpacks build failed: %w\n", err)
	}

	log.Printf("Builder: nixpacks build succeeded, image=%s", tag)
	return tag, nil
}

func ValidateRepoUrl(repoUrl string) error {
	//prevent option injection
	if strings.HasPrefix(repoUrl, "-") {
		return fmt.Errorf("invalid url")
	}

	u, err := url.Parse(repoUrl)

	if err != nil {
		return err
	}

	if u.Scheme != "https" {
		return fmt.Errorf("only https url are allowed")
	}

	if u.Host != "github.com" {
		return  fmt.Errorf("only github.com allowed")
	}

	if u.User != nil {
		return fmt.Errorf("credentials are not allowed in url")
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("query parameters and fragments are not allowed")
	}

	path := strings.TrimSuffix(strings.Trim(u.Path, "/"), ".git")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		return fmt.Errorf("repository URL must be in the form https://github.com/owner/repo")
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid repository path")
	}

	return nil

}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}