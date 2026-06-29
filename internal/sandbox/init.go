package sandbox

import (
	"fmt"

	"github.com/asim9115/containerix/internal/cgroup"

)

//create a sandbox type environment to run containers with given resources
func Init(name string, cpu float64, memory string) (*Sandbox, error) {
	sandbox := &Sandbox{
		Name: name,
		Cpu: cpu,
		Memory: memory,
	}
	fmt.Println("Intializing Sandbox")
	fmt.Println("Name:", sandbox.Name)
	fmt.Println("CPU:", sandbox.Cpu)
	fmt.Println("RAM:", sandbox.Memory)

	//create cgroup
	err := cgroup.Create(name, cpu, memory)
	if err != nil {
		return nil, err
	}
	return sandbox, nil
}



