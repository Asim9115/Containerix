package cgroup

import(
	"fmt"
	"os"
	"path/filepath"

)

const cgroupRoot = "/sys/fs/cgroup" 

func Create(name string, cpu float64, memory string) error {
	path := filepath.Join(cgroupRoot, name)

	//create cgroup directory
	if err := os.Mkdir(path, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("create cgroup %w", err)
	}

	//set memory limit
	if err := os.WriteFile(
		filepath.Join(path, "memory.max"),
		[]byte(memory),
		0644,
	); err != nil {
		return fmt.Errorf("memory limit %w", err)
	}

	quota := cpu * 100000
	cpuMax := fmt.Sprintf("%.0f 100000", quota)
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