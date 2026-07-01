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
		return fmt.Errorf("update memory limit %w", err)
	}

    // CPU limit
    var cpuMax string
    if cpu <= 0 {
        cpuMax = "max 100000"
    } else {
        quota := int64(cpu * 100000)
        cpuMax = fmt.Sprintf("%d 100000", quota)
    }

    if err := os.WriteFile(
        filepath.Join(path, "cpu.max"),
        []byte(cpuMax),
        0644,
    ); err != nil {
        return fmt.Errorf("update CPU limit: %w", err)
    }

	return nil

}