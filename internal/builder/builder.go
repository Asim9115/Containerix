package builder

import (
	"fmt"
	"log"

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


func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
