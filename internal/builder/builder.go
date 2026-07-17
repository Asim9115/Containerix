package builder

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"net/url"
	"github.com/asim9115/containerix/internal/detector"
	"github.com/asim9115/containerix/internal/docker"
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

func BuildDockerImage(temporaryPath string, detected detector.DetectResult) (string, error) {
	id := uuid.New()
	tag := "containerix-" + id.String()

	//check if docker file already exists
	if detected.HasDockerfile {
		log.Printf("Builder - Using existing Dockerfile, building image %s", tag)
		buildCommand := exec.Command("docker", "build", "-t", tag, temporaryPath)
		buildCommand.Stdout = os.Stdout
		buildCommand.Stderr = os.Stderr
		if err := buildCommand.Run(); err != nil {
			log.Printf("Builder Error - Docker build failed for %s: %v", tag, err)
			return "", fmt.Errorf("error building docker image: %w", err)
		}
		fmt.Printf("Successfully built image %s\n", tag)
		return tag, nil
	}

	var content string

	//Generate Dockerfile content based on detected language
	switch detected.Language {
	case detector.LangNode:
		var err error
		content, err = docker.GenerateNode(detected)
		if err != nil {
			return "", fmt.Errorf("node dockerfile generation failed: %w", err)
		}
	case detector.LangPython:
		var err error
		content, err = docker.GeneratePython(detected)
		if err != nil {
			return "", fmt.Errorf("python dockerfile generation failed: %w", err)
		}
	case detector.LangGo:
		var err error
		content, err = docker.GenerateGo(detected)
		if err != nil {
			return "", fmt.Errorf("go dockerfile generation failed: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported language: %q — could not detect language from repo", detected.Language)
	}

	//write generated Dockerfile
	dockerfilepath := filepath.Join(temporaryPath, "Dockerfile")
	if err := os.WriteFile(dockerfilepath, []byte(content), 0644); err != nil {
		log.Printf("Builder Error - Failed to write Dockerfile: %v", err)
		return "", err
	}
	fmt.Println("file created")
	err := os.WriteFile(dockerfilepath, []byte(content), 0644)
	if err != nil {
		log.Printf("Builder Error - Failed to write Dockerfile content: %v", err)
		return "", err
	}
	fmt.Println("wrote dockerfile")
	fmt.Println("building now")
	//building docker file
	buildCommand := exec.Command("docker", "build", "-t", tag, temporaryPath)
	buildCommand.Stdout = os.Stdout
	buildCommand.Stderr = os.Stderr
	err = buildCommand.Run()
	if err != nil {
		log.Printf("Builder Error - Docker build failed for %s: %v", tag, err)
		return "", fmt.Errorf(
			"error building docker image: %w\n%s",
			err,
			string(tag),
		)
	}
	fmt.Printf("Successfully built image %s\n", tag)
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