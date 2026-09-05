package pipeline

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/repository"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
)

func (h *State) Deploy(userId string, jobId string, logBus *types.LogBus, req *types.BuildRequest) (string, error) {

	emit := func(msg string) {
		select {
		case logBus.Ch <- types.SSEEvent{Event: "log", Data: msg}:
		default:
			<-logBus.Ch //removes the oldest message if full
			logBus.Ch <- types.SSEEvent{Event: "log", Data: msg}
		}
	}

	handleFailure := func(err error) (string, error) {
		log.Printf("Pipeline Error - Deploy failed: %v", err)
		if errUpdate := h.Repo.Deployments.UpdateError(jobId, types.DeployFailed, err.Error()); errUpdate != nil {
			log.Printf("[pipeline] failed to update error in DB: %v", errUpdate)
		}
		if errJob := h.Repo.Jobs.SetFailed(jobId, err.Error()); errJob != nil {
			log.Printf("[pipeline] failed to mark job failed in DB: %v", errJob)
		}
		emit("error: " + err.Error())
		return "", err
	}

	if req.Tier == "" {
		req.Tier = types.Tier1.Name
		req.ResolvedTier = types.Tier1
	}

	cpu := req.ResolvedTier.Cpu
	memory, err := types.MemoryToBytes(req.ResolvedTier.Memory)
	if err != nil {
		memory = "524288000"
	}
	//--------------2. Check sandbox resources------------------

	emit("checking sandbox resources")
	log.Print("checking sandbox resources")
	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "checking sandbox resources")
	err = state.SB.Sandbox.CanAllocate(cpu, memory)
	if err != nil {
		return handleFailure(err)
	}

	err = state.SB.Sandbox.Allocate(cpu, memory)
	if err != nil {
		return handleFailure(err)
	}
	allocatedSandbox := true

	cleanup := func() {
		if allocatedSandbox {
			_ = state.SB.Sandbox.Release(cpu, memory)
		}
	}

	emit("validating url : " + req.Url)
		//----------------3. Validate url from url injection-------------

	log.Printf("validating url : %s", req.Url)
	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "validating")
	if err := builder.ValidateRepoUrl(req.Url); err != nil {
		cleanup()
		return handleFailure(fmt.Errorf("validate: %w", err))
	}
	if err := builder.ValidateBuildRequest(req); err != nil {
		cleanup()
		return handleFailure(fmt.Errorf("validate: %w", err))
	}
	//----------------4. Clone the repository-------------------
	log.Printf("Cloning Repo : %s", req.Url)
	emit("cloning repository...")
	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "cloning repository")
	path, err := builder.CloneRepository(req.Url)
	if err != nil {
		cleanup()
		return handleFailure(fmt.Errorf("clone: %w", err))
	}
	defer os.RemoveAll(path)


	//docker file and container 
	containerPort := req.Port
	if containerPort <= 0 {
		containerPort = types.DefaultAppPort
	}
	tag, err := builder.BuildDockerImage(logBus , req, path)
	if err != nil {
		cleanup()
		return handleFailure(fmt.Errorf("builder: %w", err))
	}
	//-----------------8. Check internal free port ----------------------

	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "allocating host port")
	hostPort, err := state.SB.Ports.GetFreePortAndReserve()
	if err != nil {
		cleanup()
		return handleFailure(fmt.Errorf("port: %w", err))
	}
	log.Printf("Free Port : %d", hostPort)
	portReserved := true

	env := mergePlatformEnv(containerPort, req.Env)
	//-----------------9. Prepare container config----------------------

	cfg := types.Config{
		Name:  tag,
		Image: tag,
		Tier:  req.ResolvedTier,
		Env:   env,
		Ports: []types.PortMapping{
			{HostPort: hostPort, ContainerPort: containerPort},
		},
	}
	log.Printf("config : %v", cfg)
	//---------------10. Reserve the port------------------

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
			state.SB.Ports.MarkFree(hostPort)
			_ = h.Repo.Ports.FreePort(hostPort)
		}
		cleanup()
	}
	//------------11. Start the container-------------

	log.Println("Starting Container")
	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "starting container")
	cfg, err = container.Run(cfg)
	if err != nil {
		cleanupWithPort()
		return handleFailure(fmt.Errorf("start: %w", err))
	}

	cleanupWithContainer := func() {
		_ = docker.StopContainer(cfg.Name)
		_ = docker.DeleteContainer(cfg.Name)
		cleanupWithPort()
	}

	emit("waiting for app to become ready...")
	_ = h.Repo.Jobs.UpdateStatus(jobId, types.JobBuilding, "waiting-ready")
	// if err := readiness.WaitReady(cfg.Name, hostPort, req.HealthCheckPath, readiness.DefaultTimeout, emit); err != nil {
	// 	cleanupWithContainer()
	// 	return handleFailure(err)
	// }
	time.Sleep(5 * time.Second)
	//--------------12. Get pid of the container to add in cgroup------------
	pid, err := docker.GetPid(cfg.Name)
	if err != nil {
		cleanupWithContainer()
		return handleFailure(err)
	}
	log.Printf("container pid: %d", pid)
	_ = docker.DeleteImage(tag)
	//------------13. Add pid to cgroup procs-------------------

	if err := cgroup.AddProcess(state.SB.Sandbox.GetState().Name, pid); err != nil {
		cleanupWithContainer()
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

	err = h.Repo.Deployments.UpdateStatus(jobId, types.DeployRunning, cfg.Name, tag, hostPort, containerPort)
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

func mergePlatformEnv(containerPort int, user map[string]string) map[string]string {
	out := map[string]string{
		"PORT": strconv.Itoa(containerPort),
		"HOST": "0.0.0.0",
	}
	for k, v := range user {
		out[k] = v
	}
	// Platform owns PORT so publish mapping always matches.
	out["PORT"] = strconv.Itoa(containerPort)
	out["HOST"] = "0.0.0.0"
	return out
}
