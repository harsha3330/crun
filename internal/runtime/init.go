package runtime

import (
	"fmt"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	"github.com/harsha3330/crun/internal/pkg"
)

func Init(cfg config.Config) error {
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
		if err := pkg.EnsurePath(dir, true); err != nil {
			return err
		}
	}

	return nil
}
