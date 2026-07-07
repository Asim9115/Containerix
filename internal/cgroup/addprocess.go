package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func AddProcess(sandboxName string, pid int) error {
	// path: /sys/fs/cgroup/<sandboxName>/cgroup.procs
	path := filepath.Join(cgroupRoot, sandboxName, "cgroup.procs")

	processId := strconv.Itoa(pid)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cgroup.procs for sandbox %q: %w", sandboxName, err)
	}
	defer file.Close()

	if _, err := file.WriteString(processId); err != nil {
		return fmt.Errorf("failed to write pid %d to cgroup: %w", pid, err)
	}

	return nil
}