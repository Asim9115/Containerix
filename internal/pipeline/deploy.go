package pipeline

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/detector"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
)

func (h *State) Deploy(jobId string, logBus *types.LogBus, url string, tier types.Tier, env map[string]string) (string, error) {

	emit := func(msg string) {
		// non-blocking send so a stalled client never freezes the pipeline
		select {
		case logBus.Ch <- types.SSEEvent{Event: "log", Data: msg}:
		default: 
			<-logBus.Ch
			logBus.Ch <- types.SSEEvent{Event: "log", Data: msg}
		}
	}

	//-----------1. Create deployment record in DB---------
	var envJSON []byte
	if env != nil {
		envJSON, _ = json.Marshal(env)
	} else {
		envJSON = []byte("{}")
	}

	deployment := &repository.Deployment{
		ID:         jobId,
		UserID:     "test",
		RepoURL:    url,
		Status:     "building",
		TierName:   tier.Name,
		TierCPU:    tier.Cpu,
		TierMemory: tier.Memory,
		EnvJSON:    string(envJSON),
	}
	if err := h.Repo.Deployments.Create(deployment); err != nil {
		log.Printf("[pipeline] failed to create deployment record: %v", err)
	}

	// Helper function to handle failures
	handleFailure := func(err error) (string, error) {
		log.Printf("Pipeline Error - Deploy failed: %v", err)
		if errUpdate := h.Repo.Deployments.UpdateError(jobId, "failed", err.Error()); errUpdate != nil {
			log.Printf("[pipeline] failed to update error in DB: %v", errUpdate)
		}
		emit("error: " + err.Error())
		return "", err
	}

	cpu := tier.Cpu
	memory, err := types.MemoryToBytes(tier.Memory)
	if err != nil {
		memory = "524288000"
	}

	emit("checking sandbox resources")
	//--------------2. Check sandbox resources------------------
	log.Print("checking sandbox resources")
	err = state.SB.Sandbox.CanAllocate(cpu, memory)
	if err != nil {
		return handleFailure(err)
	}

	state.SB.Sandbox.Allocate(cpu, memory)
	allocatedSandbox := true

	// Cleanup callback if we fail after this point
	cleanup := func() {
		if allocatedSandbox {
			_ = state.SB.Sandbox.Release(cpu, memory)
		}
	}

	emit("validating url : " + url)
	//----------------3. Validate url from url injection-------------
	log.Printf("validating url : %s", url)
	if err := builder.ValidateRepoUrl(url); err != nil {
		cleanup()
		return handleFailure(err)
	}

	//----------------4. Clone the repository-------------------
	log.Printf("Cloning Repo : %s", url)
	emit("cloning repository...")
	path, err := builder.CloneRepository(url)
	if err != nil {
		cleanup()
		return handleFailure(err)
	}
	defer os.RemoveAll(path)

	//---------------5. Build Docker Image---------------------
	log.Printf("Building Docker image")
	emit("Building Docker image...")
	tag, err := builder.BuildDockerImage(path)
	if err != nil {
		cleanup()
		return handleFailure(err)
	}

	//---------------6. Probe to detect active container port--------------
	probeName := tag + "-probe"
	log.Printf("Running probe container %s to detect port", probeName)
	err = docker.RunContainerWithoutPorts(types.Config{
		Image: tag,
		Tier:  tier,
	}, probeName)
	if err != nil {
		cleanup()
		return handleFailure(err)
	}

	//----------------7. Get container port---------------
	ip, _ := docker.GetContainerIp(probeName)
	containerPort, err := detector.ScanActivePort(ip)
	if err != nil {
		log.Printf("Pipeline Error - Failed to determine exposed port dynamically: %v", err)
		containerPort = 3000 // Fallback
	}
	log.Printf("Dynamically Detected Container Port: %d", containerPort)
	docker.StopContainer(probeName)
	docker.DeleteContainer(probeName)

	//-----------------8. Check internal free port ----------------------
	hostPort, err := state.SB.Ports.GetFreePort()
	if err != nil {
		cleanup()
		return handleFailure(err)
	}
	log.Printf("Free Port : %d", hostPort)

	//-----------------9. Prepare container config----------------------
	cfg := types.Config{
		Name:  tag,
		Image: tag,
		Tier:  tier,
		Env:   env,
		Ports: []types.PortMapping{
			{HostPort: hostPort, ContainerPort: containerPort},
		},
	}
	log.Printf("config : %v", cfg)

	//---------------10. Reserve the port------------------
	state.SB.Ports.Reserve(cfg.Name, hostPort, containerPort)
	portReserved := true
	
	err = h.Repo.Ports.Create(&repository.Ports{
		HostPort:      hostPort,
		ContainerID:   cfg.Name,
		ContainerPort: containerPort,
	})
	if err != nil {
		log.Printf("[pipeline] failed to record port allocation in DB: %v", err)
	}

	cleanupWithPort := func() {
		if portReserved {
			state.SB.Ports.ReleasePort(hostPort)
			_ = h.Repo.Ports.FreePort(hostPort)
		}
		cleanup()
	}

	//------------11. Start the container-------------
	log.Println("Starting Container")
	cfg, err = container.Run(cfg)
	if err != nil {
		cleanupWithPort()
		return handleFailure(err)
	}

	//--------------12. Get pid of the container to add in cgroup------------
	pid, err := docker.GetPid(cfg.Name)
	if err != nil {
		_ = docker.StopContainer(cfg.Name)
		cleanupWithPort()
		return handleFailure(err)
	}
	log.Printf("container pid: %d", pid)
	docker.DeleteImage(cfg.Name)

	//------------13. Add pid to cgroup procs-------------------
	if err := cgroup.AddProcess(state.SB.Sandbox.GetState().Name, pid); err != nil {
		_ = docker.StopContainer(cfg.Name)
		cleanupWithPort()
		return handleFailure(err)
	}
	log.Printf("process %d added to cgroup", pid)

	//--------------14. Add container to sandbox------------
	state.SB.Sandbox.AddContainer(&types.Container{
		ID:     cfg.Name,
		CPU:    cfg.Tier.Cpu,
		Memory: cfg.Tier.Memory,
	})

	//-----------15. Update container status in database---------
	err = h.Repo.Deployments.UpdateStatus(jobId, "running", cfg.Name, tag, hostPort, containerPort)
	if err != nil {
		log.Printf("[pipeline] error updating status in DB: %v", err)
	}

	//-------------16. Container Url-------------------
	appUrl := fmt.Sprintf("http://localhost:%d", hostPort)
	emit_event := func(event, data string) {
		select {
		case logBus.Ch <- types.SSEEvent{Event: event, Data: data}:
		default:
		}
	}
	emit_event("deployed", appUrl)

	//------17. Return container id----------
	return cfg.Name, nil
}

