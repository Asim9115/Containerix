package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/asim9115/containerix/internal/builder/templates"
	"github.com/asim9115/containerix/internal/types"
	"github.com/google/uuid"
)

func BuildDockerImage(logbus *types.LogBus, data *types.BuildRequest, path string) (string, error) {
	id := uuid.New()
	tag := "ctx-" + id.String()[:8]
	if data.Language != types.Docker {
		var dockerFileContent string
		switch data.Language {
		case types.Python:
			dockerFileContent = templates.GeneratePython(data.BuildCommand, data.RootDirectory, data.StartCommand, data.Port)

		case types.Node:
			dockerFileContent = templates.GenerateNode(data.BuildCommand, data.RootDirectory, data.StartCommand, data.Port)

		case types.Go:
			dockerFileContent = templates.GenerateGo(data.BuildCommand, data.RootDirectory, data.StartCommand, data.Port)

		default:
			return "", fmt.Errorf("language not supported")
		}
		dockerFilePath := filepath.Join(path, "Dockerfile")
		err := os.WriteFile(dockerFilePath, []byte(dockerFileContent), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write dockerfile : %w", err)
		}

	}

	cmd := exec.Command("docker", "build", "-t", tag, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}
	return tag, nil
}
