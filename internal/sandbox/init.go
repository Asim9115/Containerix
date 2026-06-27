package sandbox

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/asim9115/containerix/internal/config"
)

//create a sandbox type environment to run containers with given resources
func Init(name string, cpu int, memory string) (*config.Config, error) {
	sb := &config.Config{
		Name: name,
		CPU: cpu,
		Memory: memory,
	}
	fmt.Println("Intializing Sandbox")
	fmt.Println("Name:", sb.Name)
	fmt.Println("CPU:", sb.CPU)
	fmt.Println("RAM:", sb.Memory)

	//create cgroup
	err := Create(sb)
	if err != nil {
		return nil, err
	}
	return sb, nil
}

const cgroupRoot = "/sys/fs/cgroup" 

func Create(cgroup *config.Config) error {
	path := filepath.Join(cgroupRoot, cgroup.Name)

	//create cgroup directory
	if err := os.Mkdir(path, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("create cgroup %w", err)
	}

	//set memory limit
	if err := os.WriteFile(
		filepath.Join(path, "memory.max"),
		[]byte(cgroup.Memory),
		0644,
	); err != nil {
		return fmt.Errorf("memory limit %w", err)
	}

	quota := cgroup.CPU * 100000
	cpuMax := fmt.Sprintf("%d 100000", quota)
	//set cpu limit
	if err := os.WriteFile(
		filepath.Join(path, "cpu.max"),
		[]byte(cpuMax),
		0644,
	); err != nil {
		return fmt.Errorf("Cpu limit: %w", err)
	}

	return nil
}