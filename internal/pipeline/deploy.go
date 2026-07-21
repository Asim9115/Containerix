package pipeline

import (
	"fmt"
	"log"
	"os"

	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/detector"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
)

func Deploy(jobId string, logBus *types.LogBus, url string) (string, error) {
	emit := func(msg string) {
		// non-blocking send so a stalled client never freezes the pipeline
		select {
		case logBus.Ch <- types.SSEEvent{Event: "log", Data: msg}:
		default: // drop if buffer full — pipeline keeps going
		}
	}
	//temporary values
	cpu := types.Tier1.Cpu
	memory := "524288000"
	emit("checking sandbox resources")
	//1. Check sandbox resources
	log.Print("checking sandbox resources")
	err := state.SB.Sandbox.CanAllocate(cpu, memory)
	if err != nil {
		log.Printf("Pipeline Error - Sandbox allocation failed: %v", err)
		return "", err
	}
	state.SB.Sandbox.Allocate(cpu, memory)
	emit("validating url : " + url)
	log.Printf("validating url : %s", url)

	//2. Validate url from url injection
	if err := builder.ValidateRepoUrl(url); err != nil {
		log.Printf("Pipeline Error - Invalid URL %s: %v", url, err)
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	log.Printf("Cloning Repo : %s", url)
	emit("cloning repository...")
	//3. Clone the repository
	path, err := builder.CloneRepository(url)
	if err != nil {
		log.Printf("Pipeline Error - Repository clone failed for %s: %v", url, err)
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	defer os.RemoveAll(path)

	log.Printf("Building Docker image")
	emit("Building Docker image...")

	log.Printf("Building Docker image")
	//5. Build Docker Image
	tag, err := builder.BuildDockerImage(path)
	if err != nil {
		log.Printf("Pipeline Error - Docker build failed: %v", err)
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}

	//6. Probe to detect active container port
	probeName := tag + "-probe"
	log.Printf("Running probe container %s to detect port", probeName)

	err = docker.RunContainerWithoutPorts(types.Config{
		Image: tag,
		Tier:  types.Tier1,
	}, probeName)

	if err != nil {
		// handle probe run failure
		log.Printf("Pipeline Error - Probe run failed: %v", err)
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	ip, _ := docker.GetContainerIp(probeName)
	containerPort, err := detector.ScanActivePort(ip)

	if err != nil {
		log.Printf("Pipeline Error - Failed to determine exposed port dynamically: %v", err)
		containerPort = 3000 // Fallback
	}
	log.Printf("Dynamically Detected Container Port: %d", containerPort)

	// Cleanup Probe Container
	docker.StopContainer(probeName)
	docker.DeleteContainer(probeName)

	//6. get free port
	hostPort, err := state.SB.Ports.GetFreePort()
	if err != nil {
		log.Printf("Pipeline Error - Port allocation failed: %v", err)
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	log.Printf("Free Port : %d", hostPort)

	cfg := types.Config{
		Name:  tag,
		Image: tag,
		Tier:  types.Tier1,
		Ports: []types.PortMapping{
			{HostPort: hostPort, ContainerPort: containerPort},
		},
	}
	log.Printf("config : %v", cfg)
	//10. mark port as used
	state.SB.Ports.Reserve(cfg.Name, hostPort, containerPort)

	//11. Update sandbox resources

	log.Println("Starting Container")

	// 7. Run the container
	cfg, err = container.Run(cfg)
	if err != nil {
		log.Printf("Pipeline Error - Container run failed: %v", err)
		state.SB.Sandbox.Release(cpu, memory)
		state.SB.Ports.ReleasePort(hostPort)
		return "", err
	}

	// 8. Get PID of the running container by its name (not image tag)
	pid, err := docker.GetPid(cfg.Name)
	if err != nil {
		_ = docker.StopContainer(cfg.Name)
		state.SB.Sandbox.Release(cfg.Tier.Cpu, cfg.Tier.Memory)
		state.SB.Ports.ReleasePort(hostPort)
		log.Printf("Pipeline Error - Failed to get PID for %s: %v", cfg.Name, err)
		return "", err
	}
	log.Printf("container pid: %d", pid)

	// 9. Add container process to sandbox cgroup
	if err := cgroup.AddProcess(state.SB.Sandbox.GetState().Name, pid); err != nil {
		_ = docker.StopContainer(cfg.Name)
		state.SB.Sandbox.Release(cfg.Tier.Cpu, cfg.Tier.Memory)
		state.SB.Ports.ReleasePort(hostPort)
		log.Printf("Pipeline Error - Failed to add process %d to cgroup: %v", pid, err)
		return "", err
	}
	log.Printf("process %d added to cgroup", pid)

	//12. store container in sandbox state cleanly
	state.SB.Sandbox.AddContainer(&types.Container{
		ID:     cfg.Name,
		Config: cfg,
		Status: "running",
	})

	appUrl := fmt.Sprintf("http://localhost:%d", hostPort)
	emit_event := func(event, data string) {
		select {
		case logBus.Ch <- types.SSEEvent{Event: event, Data: data}:
		default:
		}
	}
	emit_event("deployed", appUrl)
	//13. Return container id and host port
	return cfg.Name, nil
}
