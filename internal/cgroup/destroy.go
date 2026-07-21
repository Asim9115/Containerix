package cgroup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/asim9115/containerix/internal/container"
	"github.com/asim9115/containerix/internal/types"
)

// Destroy stops all containers then removes the cgroup leaf directory.

func Destroy(name string, rootpath string, containers map[string]*types.Container) error {
	log.Printf("[cgroup.Destroy] Stopping all containers in sandbox %q", name)
	container.StopAll(containers)
	log.Printf("[cgroup.Destroy] All containers stopped, removing cgroup directory")
	path := filepath.Join(rootpath, name)
	log.Printf("[cgroup.Destroy] Removing cgroup path: %s", path)


	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			log.Printf("[cgroup.Destroy] cgroup path %q already gone, skipping", path)
			return nil
		}
		return fmt.Errorf("cgroup.Destroy: failed to remove %q: %w", path, err)
	}

	log.Printf("[cgroup.Destroy] cgroup %q removed successfully", name)
	return nil
}