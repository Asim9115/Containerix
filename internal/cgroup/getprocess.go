package cgroup

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func GetProcesses() ([]string, error){
	path := filepath.Join(cgroupRoot,"/containerix/cgroup.procs")

	file, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		log.Printf("error opening file : %v",err)
		return nil, err
	}
	defer file.Close()

	var processes []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		processes = append(processes, scanner.Text())
	}
	log.Printf("processes : %v", processes)
	if err = scanner.Err(); err != nil {
		return  nil, fmt.Errorf("error reading cgroup file: %w", err)
	}
	return processes, nil
}