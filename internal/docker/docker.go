package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/asim9115/containerix/internal/types"
)
func RunContainer(cfg types.Config) error {
	if len(cfg.Ports) == 0{
		return fmt.Errorf("RunContainer: cfg.Ports is empty")
	}
	port := fmt.Sprintf("%d:%d", cfg.Ports[0].HostPort, cfg.Ports[0].ContainerPort)

	args := []string{
		"run", "-d",
        "--name", cfg.Name,
        "-p", port,
        "--cpus", strconv.FormatFloat(cfg.Tier.Cpu, 'f', -1, 64),
        "--memory", cfg.Tier.Memory,
        "--memory-swap", cfg.Tier.Memory,
        "--pids-limit", strconv.Itoa(cfg.Tier.PidsLimit),
        "--security-opt", "no-new-privileges",
	}

	for key, value := range cfg.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value) )
	}

   args = append(args, cfg.Image)
   
    cmd := exec.Command("docker", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("failed to start container: %s (err: %w)", string(output), err)
    }
    log.Printf("Container started successfully! ID: %s", strings.TrimSpace(string(output)))
    return nil

}

func StopContainer(id string) error {
	
	cmd := exec.Command("docker", "stop", id)

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to stop container: %s", string(output))
	}

	return nil
}

func StartContainer(id string) error {
	cmd := exec.Command("docker", "start", id)
	_, err := cmd.CombinedOutput()

	if err!= nil {
		return fmt.Errorf("failed to start container %s", id)
	}
	log.Printf("container %s started successfully", id)
	return nil
}

func GetPid(tag string) (int, error) {

	output, err := exec.Command("docker", "inspect", "-f", "{{.State.Pid}}", tag).Output()
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(output))) 
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func DeleteContainer(id string) error {
    out, err := exec.Command("docker", "rm", id).CombinedOutput()
    if err != nil {
        return fmt.Errorf("docker rm failed: %v: %s", err, out)
    }
    return nil
}

func DeleteImage(id string) error {
    out, err := exec.Command("docker", "rmi", id).CombinedOutput()
    if err != nil {
        return fmt.Errorf("docker rmi failed: %v: %s", err, out)
    }
    return nil
}

// ForceRemoveContainer stops (if needed) and force-removes a container,
// ignoring "no such container" errors so probe cleanup never blocks the pipeline.
func ForceRemoveContainer(id string) {
	// best-effort stop; ignore errors (container may have already exited)
	_ = exec.Command("docker", "stop", id).Run()
	out, err := exec.Command("docker", "rm", "-f", id).CombinedOutput()
	if err != nil {
		outStr := strings.TrimSpace(string(out))
		// "No such container" is fine — it was already gone
		if !strings.Contains(strings.ToLower(outStr), "no such container") {
			log.Printf("[docker] ForceRemoveContainer %s: %v — %s", id, err, outStr)
		}
	}
}

// GetExposedPorts returns the list of ports declared via EXPOSE in the image
// (or in the container's Config if the container has already been created).
// It inspects the *image*, so it works before a container is started.
// Returns ports as integers; the protocol suffix ("/tcp") is stripped.
func GetExposedPorts(imageTag string) ([]int, error) {
	out, err := exec.Command(
		"docker", "inspect",
		"--type", "image",
		"--format", "{{json .Config.ExposedPorts}}",
		imageTag,
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect image %q: %w", imageTag, err)
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "null" || trimmed == "" {
		return nil, nil // image has no EXPOSE directive
	}

	// ExposedPorts is map[string]struct{} serialised as e.g. {"3000/tcp":{}}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &raw); err != nil {
		return nil, fmt.Errorf("parse ExposedPorts: %w", err)
	}

	ports := make([]int, 0, len(raw))
	for key := range raw {
		// key format: "3000/tcp", "8080/udp", etc.
		parts := strings.SplitN(key, "/", 2)
		p, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		ports = append(ports, p)
	}
	return ports, nil
}

func RunContainerWithoutPorts(cfg types.Config, probeName string) error {
	// NOTE: --read-only is intentionally omitted here.
	// Many frameworks (Django, Rails, etc.) write .pyc / tmp files on startup.
	// A read-only root FS causes the process to crash before it can bind a port,
	// making port detection impossible. We use --tmpfs instead to keep /tmp writable
	// while everything else stays inside the container's layered FS.
	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"--name", probeName,
		"--cpus", strconv.FormatFloat(cfg.Tier.Cpu, 'f', -1, 64),
		"--memory", cfg.Tier.Memory,
		"--memory-swap", cfg.Tier.Memory,
		"--pids-limit", strconv.Itoa(cfg.Tier.PidsLimit),
		"--security-opt", "no-new-privileges",
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=64m",
		cfg.Image,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start probe container: %s (err: %w)", string(output), err)
	}
	return nil
}

func GetContainerIp(id string) (string, error) {
	output, err := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", id).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// StreamContainerLogs attaches to a running container and forwards every log
// line to outCh.  Returns when the container stops or ctx is cancelled.
func StreamContainerLogs(ctx context.Context, containerName string, outCh chan<- types.SSEEvent) error {
    cmd := exec.CommandContext(ctx, "docker", "logs", "-f", "--tail", "50", containerName)

    pr, pw := io.Pipe()
    cmd.Stdout = pw
    cmd.Stderr = pw

    if err := cmd.Start(); err != nil {
        pw.Close()
        return err
    }

    scanner := bufio.NewScanner(pr)
    for scanner.Scan() {
        select {
        case outCh <- types.SSEEvent{Event: "log", Data: scanner.Text()}:
        case <-ctx.Done():
            pw.Close()
            cmd.Process.Kill()
            return ctx.Err()
        }
    }
    pw.Close()
    return cmd.Wait()
}
