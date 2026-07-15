package cgroup

import (
	"log"
	"os"
	"path/filepath"
    "github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/types"
)

// Destroy stops all containers then removes the cgroup directory.
func Destroy(name string, rootpath string, containers map[string]*types.Container) error {
	log.Println("destroy cgroup : stopping all containers")
	container.StopAll(containers)
	path := filepath.Join(rootpath, name)
	return os.RemoveAll(path)
}