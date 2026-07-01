package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
)

func Destroy(name string, rootpath string) error {
	path := filepath.Join(rootpath, name)
	err := os.Remove(path)

	if err != nil {
		return fmt.Errorf("Failed to delete cgroup {%w}\n",err)
	}
	return nil
}