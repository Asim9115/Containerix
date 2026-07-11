package pipeline

import (
	"github.com/asim9115/containerix/internal/builder"
	"github.com/asim9115/containerix/internal/cgroup"
	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/detector"
	"github.com/asim9115/containerix/internal/docker"
	"github.com/asim9115/containerix/internal/state"
	"github.com/asim9115/containerix/internal/types"
	"log"
	"os"
)

func Deploy(jobId string, url string) (string, error) {

	//temporary values
	cpu := types.Tier1.Cpu
	memory := "524288000"

	//1. Check sandbox resources
	log.Print("checking sandbox resources")
	err := state.SB.Sandbox.CanAllocate(cpu, memory)
	if err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	state.SB.Sandbox.Allocate(cpu, memory)

	log.Printf("validating url : %s", url)

	//2. Validate url from url injection
	if err := builder.ValidateRepoUrl(url); err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	log.Printf("Cloning Repo : %s", url)

	//3. Clone the repository
	path, err := builder.CloneRepository(url)
	if err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	log.Print("Detecting Language")

	//4. Detect Language or DockerFile
	result := detector.Detect(path)
	log.Printf("Building Docker image")

	//5. Build Docker Image
	tag, err := builder.BuildDockerImage(path, result)
	if err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	defer os.RemoveAll(path)

	//6. get free port
	hostPort, err := state.SB.Ports.GetFreePort()
	if err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		return "", err
	}
	log.Printf("Free Port : %d", hostPort)

	cfg := types.Config{
		Name:  "containerix-" + tag,
		Image: tag,
		Tier:  types.Tier1,
		Ports: []types.PortMapping{
			{HostPort: hostPort, ContainerPort: 8000},
		},
	}
	log.Printf("config : %v", cfg)
	//10. mark port as used
	state.SB.Ports.Reserve(cfg.Name, hostPort, 8000)

	//11. Update sandbox resources

	log.Println("Starting Container")

	// 7. Run the container
	cfg, err = container.Run(cfg)
	if err != nil {
		state.SB.Sandbox.Release(cpu, memory)
		state.SB.Ports.ReleasePort(hostPort)
		return "", err
	}

	// 8. Get PID of the running container by its name (not image tag)
	pid, err := docker.GetPid(cfg.Name)
	if err != nil {
		return "", err
	}
	log.Printf("container pid: %d", pid)

	// 9. Add container process to sandbox cgroup
	if err := cgroup.AddProcess(state.SB.Sandbox.GetState().Name, pid); err != nil {
		return "", err
	}
	log.Printf("process %d added to cgroup", pid)

	//12. store container in sandbox state cleanly
	state.SB.Sandbox.AddContainer(&types.Container{
		ID:     cfg.Name,
		Config: cfg,
		Status: "running",
	})

	//13. Return container id and host port
	return cfg.Name, nil
}
