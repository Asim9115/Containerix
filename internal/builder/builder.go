package builder

import (
	"fmt"
	"os/exec"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)


func CloneRepository(url string) (string, error) {
	destPath := filepath.Join(os.TempDir(), uuid.New().String())

	fmt.Println("cloning repository:", url, "→", destPath)

	cmd := exec.Command("git", "clone", url, destPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("git clone failed: %s", string(output))
	}
	return destPath, nil

}