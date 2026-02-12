package runtime

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
	"github.com/harsha3330/crun/internal/pkg"
)

type InitOptions struct {
	LogLevel *config.LogLevel
}

func Init(cfg *config.Config, log *slog.Logger, stater logger.Console, opts *InitOptions) error {
	log.Debug("init arguments", "options", opts)

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
		} else {
			msg := fmt.Sprintf("created directory %s", dir)
			stater.Success(msg)
		}
	}

	if opts != nil && opts.LogLevel != nil {
		cfg.LogLevel = *opts.LogLevel
	}

	if err := config.Write(*cfg); err != nil {
		log.Error("failed to write config",
			"path", cfg.ConfigFilePath,
			"err", err,
		)
		stater.Error("failed to config toml file")
		return err
	}

	log.Info("crun initialized",
		"rootDir", cfg.RootDir,
		"config", cfg.ConfigFilePath,
	)
	stater.Success("created toml for config succesfully")
	return nil
}
