package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"net/url"
	"github.com/asim9115/containerix/internal/detector"
	"github.com/asim9115/containerix/internal/dockerfile"
	"github.com/google/uuid"
)


func CloneRepository(repoUrl string) (string, error) {

	//validate url

	if err := ValidateRepoUrl(repoUrl); err != nil {
		return "", err
	}

	//path to store the repository files 
	destPath := filepath.Join("tmp", uuid.New().String())

    if err := os.MkdirAll("tmp", 0755); err != nil {
        return "", err
    }
	fmt.Println("cloning repository:", repoUrl, "→", destPath)

	//executing git clone command
	cmd := exec.Command("git", "clone", repoUrl, destPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
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
		content, err = dockerfile.GenerateNode(detected)
		
	case detector.LangPython:
		content, err = dockerfile.GeneratePython(detected)
	case detector.LangGo:
		content, err = dockerfile.GenerateGo(detected)
	default:
		return "", fmt.Errorf("unsupported language: %s", detected.Language)
	}
	//creating docker file
	dockerfilepath := filepath.Join(temporaryPath, "Dockerfile")
	_, err = os.Create(dockerfilepath)
	if err != nil {
		return "", err
	}
	fmt.Println("file created")
	err = os.WriteFile(dockerfilepath, []byte(content), 0644)
	if err != nil {
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

	switch u.Scheme {
	case "https", "http", "ssh":
		return nil
	default:
		return fmt.Errorf("unsupported repository scheme: %s", u.Scheme)
	}
}