package cgroup

import (
	"fmt"
	"os"
)

func Destroy(name string, path string) error {
	err := os.RemoveAll(path)

	if err != nil {
		return fmt.Errorf("Failed to delete cgroup {%w}\n",err)
	}
	return nil
}