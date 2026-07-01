package cgroup

import (
	"path/filepath"
	"fmt"
	"os"
)

func Update(name string, cpu float64, memory string, rootpath string) error {
	path := filepath.Join(rootpath, name)

	//set memory limit
	if err := os.WriteFile(
		filepath.Join(path, "memory.max"),
		[]byte(memory),
		0644,
	); err != nil {
		return fmt.Errorf("memory limit %w", err)
	}

	quota := cpu * 100000
	cpuMax := fmt.Sprintf("%f 100000", quota)
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