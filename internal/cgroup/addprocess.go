package cgroup

import (
	"os"
	"path/filepath"
	"strconv"
)

func AddProcess(pid int) error {
	path := filepath.Join(cgroupRoot, "containerix/cgroup.procs")
	//convert int to string
	processId := strconv.Itoa(pid)

	//open write writeonly and append data to the end
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return err
	}
	//close file after all operation
	defer file.Close()
	//write the process
	if _, err := file.WriteString(processId); err != nil {
		return err
	}

	return nil
}