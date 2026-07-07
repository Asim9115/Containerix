package sandbox

import (
	"fmt"
	"github.com/asim9115/containerix/internal/types"
	"github.com/asim9115/containerix/internal/cgroup"
)

// create a sandbox type environment to run containers with given resources
func Init(name string, cpu float64, memory string) (Sandbox, error) {
	sandboxData := &types.Sandbox{
		Name:       name,
		Cpu:        cpu,
		Memory:     memory,
		UsedMemory: "0",
		UsedCpu:    0,
		Containers: make(map[string]*types.Container),
	}
	fmt.Println("Initializing Sandbox")
	fmt.Println("Name:", sandboxData.Name)
	fmt.Println("CPU:", sandboxData.Cpu)
	fmt.Println("RAM:", sandboxData.Memory)

	// create cgroup
	err := cgroup.Create(name, cpu, memory)
	if err != nil {
		return nil, err
	}
	return &SandboxManager{Sandbox: sandboxData}, nil
}
