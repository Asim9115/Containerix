package builder

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func CloneRepository(repoUrl string) (string, error) {
	destPath := filepath.Join("tmp", uuid.New().String())

	if err := os.MkdirAll("tmp", 0755); err != nil {
		log.Printf("Builder Error - Failed to create tmp directory: %v", err)
		return "", err
	}
	fmt.Println("cloning repository:", repoUrl, "→", destPath)

	cmd := exec.Command("git", "clone", "--depth", "1", repoUrl, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Builder Error - Git clone failed for %s: %v. Output: %s", repoUrl, err, string(output))
		return "", fmt.Errorf("git clone failed: %s", string(output))
	}
	fmt.Println("Clone success")
	return destPath, nil
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
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
