package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func Write(cfg Config) error {
	if cfg.ConfigFilePath == "" {
		return fmt.Errorf("config file path is empty")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(cfg.ConfigFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically
	tmp := cfg.ConfigFilePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	return os.Rename(tmp, cfg.ConfigFilePath)
}
