package proxy

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/asim9115/containerix/internal/config"
)

func AddRoute(domain, upstream string) error {
	data := fmt.Sprintf(`%s {
	reverse_proxy %s
	}`, domain, upstream)
	caddyPath := config.Load().CaddyPath
	fileName := filepath.Join(caddyPath,"containerix",domain+".caddy")

	if err := os.MkdirAll(
		filepath.Dir(fileName), 0755,
	); err != nil {
		return err
	}

	if err := os.WriteFile(
		fileName, []byte(data), 0644,
	); err != nil {
		return err
	}
	
	if err := Reload(); err != nil {
		//fallback and remove file
		if removeErr := os.Remove(fileName); removeErr != nil {
			return fmt.Errorf(
				"reload failed: %v; rollback failed: %w",
				err,
				removeErr,
			)
		}
		return fmt.Errorf("failed to reload : %w", err)
	}
	return nil
}

func Reload() error {
	caddyfile := filepath.Join(
		config.Load().CaddyPath, "caddyfile",
	)

	validate := exec.Command("caddy", "validate", "--config", caddyfile)

	if output, err := validate.CombinedOutput(); err != nil {
		return fmt.Errorf(
			"caddy vlidation failed %w: %s", err, output,
		)
	}

	reload := exec.Command(
		"caddy",
		"reload",
		"--config",
		caddyfile,
	)

	if output, err := reload.CombinedOutput(); err != nil {
		return fmt.Errorf(
			"caddy reload failed: %w: %s",
			err,
			output,
		)
	}

	return nil
}

func RemoveRoute(domain string) error {
	caddyfile := config.Load().CaddyPath
	filename := filepath.Join(caddyfile, "containerix", domain+".caddy")

	data, err := os.ReadFile(filename);
	if err != nil {
		return err
	}
	
	if err := os.Remove(filename); err != nil {
		return err
	}

	if err := Reload(); err != nil {
		//rollback and restore file
		if restoreErr := os.WriteFile(filename, data, 0644); restoreErr != nil {
			return fmt.Errorf(
				"reload failed: %v; rollback failed: %w",
				err,
				restoreErr,
			)
		}
		_ = Reload()
		return fmt.Errorf("failed to reload: %w", err)
	}
	return nil
}