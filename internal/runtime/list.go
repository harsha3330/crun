package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
)

// ImageList returns all pulled images as "repo:tag" (e.g. "nginx:1-alpine-perl").
func ImageList(cfg config.Config, stater logger.Console) ([]string, error) {
	imagesDir := filepath.Join(cfg.RootDir, "images")
	entries, err := os.ReadDir(imagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		stater.Error("failed to read images dir", "error", err)
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		repo := e.Name()
		tagsDir := filepath.Join(imagesDir, repo, "tags")
		tagEntries, err := os.ReadDir(tagsDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			continue
		}
		for _, te := range tagEntries {
			if te.IsDir() {
				continue
			}
			out = append(out, repo+":"+te.Name())
		}
	}
	return out, nil
}

// ContainerInfo holds one row for container list.
type ContainerInfo struct {
	ID     string
	Image  string
	PID    int
	Status string
}

// ContainerList returns running containers (those with a pid file).
func ContainerList(cfg config.Config, stater logger.Console) ([]ContainerInfo, error) {
	containersDir := filepath.Join(cfg.RootDir, "containers")
	entries, err := os.ReadDir(containersDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		stater.Error("failed to read containers dir", "error", err)
		return nil, err
	}
	var out []ContainerInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		id := e.Name()
		pidPath := filepath.Join(containersDir, id, "pid")
		pidData, err := os.ReadFile(pidPath)
		if err != nil {
			continue
		}
		var pid int
		if _, err := fmt.Sscanf(string(pidData), "%d", &pid); err != nil {
			continue
		}
		status := "running"
		if err := syscall.Kill(pid, 0); err != nil {
			if err == syscall.ESRCH {
				status = "exited"
			}
		}
		image := ""
		imagePath := filepath.Join(containersDir, id, "image")
		if b, err := os.ReadFile(imagePath); err == nil {
			image = string(b)
		}
		out = append(out, ContainerInfo{ID: id, Image: image, PID: pid, Status: status})
	}
	return out, nil
}
