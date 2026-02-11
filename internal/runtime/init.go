package runtime

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	"github.com/harsha3330/crun/internal/pkg"
)

func Init(cfg config.Config, log *slog.Logger) error {
	if cfg.RootDir == "" {
		return fmt.Errorf("root dir is empty")
	}

	dirs := []string{
		cfg.RootDir,
		filepath.Join(cfg.RootDir, "images"),
		filepath.Join(cfg.RootDir, "containers"),
		filepath.Join(cfg.RootDir, "layers"),
		filepath.Join(cfg.RootDir, "blobs"),
	}

	for _, dir := range dirs {
		if err := pkg.EnsureDir(dir); err != nil {
			log.Error("ensure path failed",
				"path", dir,
				"err", err,
			)
			return fmt.Errorf("init failed for path %s: %w", dir, err)
		}
	}

	return nil
}
