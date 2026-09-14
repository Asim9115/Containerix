package builder

import (
	"bufio"
	"fmt"
	"io"
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
	buildCommand := buildLine(data.BuildCommand)
	if data.Language != types.Docker {
		var dockerFileContent string
		switch data.Language {
		case types.Python:
			dockerFileContent = templates.GeneratePython(buildCommand, data.RootDirectory, data.StartCommand, data.Port)

		case types.Node:
			dockerFileContent = templates.GenerateNode(buildCommand, data.RootDirectory, data.StartCommand, data.Port)

		case types.Go:
			dockerFileContent = templates.GenerateGo(buildCommand, data.RootDirectory, data.StartCommand, data.Port)

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

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		return "", fmt.Errorf("docker build start: %w", err)
	}

	// Scanner runs in a goroutine so it doesn't block cmd.Wait().
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(pr)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			line := scanner.Text()
			select {
			case logbus.Ch <- types.SSEEvent{Event: "log", Data: line}:
			default:
				<-logbus.Ch
				logbus.Ch <- types.SSEEvent{Event: "log", Data: line}
			}
		}
	}()

	// Wait for docker build to finish then close pw so scanner sees EOF.
	buildErr := cmd.Wait()
	pw.Close()

	// Wait for scanner goroutine to drain remaining lines before returning.
	<-scanDone

	if buildErr != nil {
		return "", fmt.Errorf("docker build failed: %w", buildErr)
	}
	return tag, nil
}

func buildLine(cmd string) string {
	if cmd == "" {
		return ""
	}
	return "RUN " + cmd + "\n"
}
