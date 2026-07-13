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
		buildCommand := exec.Command("docker", "build", "-t", tag, temporaryPath)
		output, err := buildCommand.CombinedOutput()
		if err != nil {
			log.Printf("Builder Error - Docker build failed for %s: %v. Output: %s", tag, err, string(output))
			return "", fmt.Errorf(
				"error building docker image: %w\n%s",
				err,
				string(output),
			)
		}
		fmt.Printf("Successfully built image %s\n", tag)
		return tag, nil
	}

	var (
		content string
		err 	error
	)
	//Generate Dockerfile based on detected language
	switch detected.Language {
	case detector.LangNode:
		content, err = docker.GenerateNode(detected)
		
	case detector.LangPython:
		content, err = docker.GeneratePython(detected)
	case detector.LangGo:
		content, err = docker.GenerateGo(detected)
	default:
		return "", fmt.Errorf("unsupported language: %s", detected.Language)
	}
	//creating docker file
	dockerfilepath := filepath.Join(temporaryPath, "Dockerfile")
	_, err = os.Create(dockerfilepath)
	if err != nil {
		log.Printf("Builder Error - Failed to create Dockerfile: %v", err)
		return "", err
	}
	fmt.Println("file created")
	err = os.WriteFile(dockerfilepath, []byte(content), 0644)
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