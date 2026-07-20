package docker

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"github.com/asim9115/containerix/internal/types"
	"bufio"
	"context"
	"io"
)
func RunContainer(cfg types.Config) error {
	hostPort := cfg.Ports[0].HostPort
	containerPort := cfg.Ports[0].ContainerPort
	port := fmt.Sprintf("%d:%d", hostPort, containerPort)

	cmd := exec.Command(
    "docker",
    "run",
    "-d",
    "--name", cfg.Name,
    "-p", port,
    "--cpus", strconv.FormatFloat(cfg.Tier.Cpu, 'f', -1, 64),
    "--memory", cfg.Tier.Memory,
    "--memory-swap", cfg.Tier.Memory,
    "--pids-limit", strconv.Itoa(cfg.Tier.PidsLimit),
    "--security-opt", "no-new-privileges",
    "--read-only",
    cfg.Image,
)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container: %s (err: %w)", string(output), err)
	}

	log.Printf("Container started successfully! ID: %s", string(output))
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

func RunContainerWithoutPorts(cfg types.Config, probeName string) error {
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
		"--read-only",
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
