package docker

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"github.com/asim9115/containerix/internal/types"
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