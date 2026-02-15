package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
)

func PidPath(cfg config.Config, containerID string) string {
	return filepath.Join(cfg.RootDir, "containers", containerID, "pid")
}

func Stop(cfg config.Config, stater logger.Console, containerID string) error {
	pidPath := PidPath(cfg, containerID)
	data, err := os.ReadFile(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			stater.Error("container not found or not running", "container-id", containerID)
			return fmt.Errorf("container %s: no pid file (not running or invalid id)", containerID)
		}
		stater.Error("failed to read pid file", "error", err)
		return err
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		stater.Error("invalid pid file", "path", pidPath)
		return fmt.Errorf("invalid pid file: %w", err)
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		if err == syscall.ESRCH {
			stater.Warn("process already gone", "pid", pid)
			removeContainerFS(cfg, containerID, pidPath, stater)
			return nil
		}
		stater.Error("failed to send SIGTERM", "error", err)
		return err
	}

	const waitTerm = 3
	for i := 0; i < waitTerm*10; i++ {
		if err := syscall.Kill(pid, 0); err != nil {
			if err == syscall.ESRCH {
				break
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := syscall.Kill(pid, 0); err == nil {
		_ = syscall.Kill(pid, syscall.SIGKILL)
	}

	removeContainerFS(cfg, containerID, pidPath, stater)
	stater.Success("container stopped", "container-id", containerID)
	return nil
}

func removeContainerFS(cfg config.Config, containerID string, pidPath string, stater logger.Console) {
	_ = os.Remove(pidPath)

	containerDir := filepath.Join(cfg.RootDir, "containers", containerID)
	mergedPath := filepath.Join(containerDir, "merged")

	if err := syscall.Unmount(mergedPath, 0); err != nil {
		if err != syscall.EINVAL {
			stater.Warn("unmount overlay failed", "path", mergedPath, "error", err)
		}
	}

	if err := os.RemoveAll(containerDir); err != nil {
		stater.Warn("remove container dir failed", "path", containerDir, "error", err)
	}
}
