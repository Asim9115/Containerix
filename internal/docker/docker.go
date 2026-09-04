package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/asim9115/containerix/internal/types"
	"github.com/moby/moby/client"
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

// ContainerRunning reports whether the named container is currently running.
func ContainerRunning(name string) (bool, error) {
	out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("inspect %s: %s", name, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)) == "true", nil
}

// TailLogs returns the last n lines of container logs.
func TailLogs(name string, n int) (string, error) {
	if n <= 0 {
		n = 50
	}
	out, err := exec.Command("docker", "logs", "--tail", strconv.Itoa(n), name).CombinedOutput()
	return strings.TrimSpace(string(out)), err
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

func GetContainerIp(id string) (string, error) {
	output, err := exec.Command("docker", "inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", id).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// StreamContainerLogs follows a running container and calls write for each line.
// Returns when the container stops or ctx is cancelled (e.g. client disconnect).
func StreamContainerLogs(ctx context.Context, containerName string, write func(types.SSEEvent) error) error {
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", "--tail", "100", containerName)

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		return err
	}

	go func() {
		<-ctx.Done()
		pw.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}()

	scanner := bufio.NewScanner(pr)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := write(types.SSEEvent{Event: "log", Data: scanner.Text()}); err != nil {
			return err
		}
	}
	pw.Close()
	return cmd.Wait()
}

// containerIDFromCgroup reads /proc/<pid>/cgroup to extract the 64-char Docker
// container ID embedded in the cgroup path. Works for every thread in the
// container (not just the init process), since all threads share the same cgroup.
// Returns "" without error if the PID doesn't belong to a Docker container.
func containerIDFromCgroup(pid string) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/cgroup", pid))
	if err != nil {
		return ""
	}
	// Docker embeds the full 64-char container ID in the cgroup path, e.g.:
	//   cgroup v2:  0::/system.slice/docker-<ID>.scope
	//   cgroup v1:  12:devices:/docker/<ID>
	re := regexp.MustCompile(`[a-f0-9]{64}`)
	for _, line := range strings.Split(string(data), "\n") {
		if match := re.FindString(line); match != "" {
			return match
		}
	}
	return ""
}

func GetContainerIDFromPID(pid string) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("docker client init: %w", err)
	}
	defer cli.Close()
	ctx := context.Background()

	// Strategy 1: read the 64-char container ID from /proc/<pid>/cgroup.
	// This works for ALL threads in a container (not just the init PID).
	if fullID := containerIDFromCgroup(pid); fullID != "" {
		log.Printf("[docker] pid %s -> cgroup container ID %s", pid, fullID[:12])
		inspect, err := cli.ContainerInspect(ctx, fullID, client.ContainerInspectOptions{})
		if err == nil && len(inspect.Container.Name) > 0 {
			return strings.TrimPrefix(inspect.Container.Name, "/"), nil
		}
	}

	// Strategy 2 (fallback): match State.Pid against the given PID.
	// Only matches the container init process, not its child threads.
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return "", nil
	}
	result, err := cli.ContainerList(ctx, client.ContainerListOptions{All: false})
	if err != nil {
		return "", fmt.Errorf("container list: %w", err)
	}
	for _, c := range result.Items {
		inspect, err := cli.ContainerInspect(ctx, c.ID, client.ContainerInspectOptions{})
		if err != nil {
			continue
		}
		if inspect.Container.State != nil && inspect.Container.State.Pid == pidInt {
			if len(c.Names) > 0 {
				return strings.TrimPrefix(c.Names[0], "/"), nil
			}
			return c.ID, nil
		}
	}
	return "", nil
}