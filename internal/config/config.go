package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	logger "github.com/harsha3330/crun/internal/log"
)

type Config struct {
	RootDir        string
	AppLogDir      string
	ConfigFilePath string
	LogLevel       logger.LogLevel
	LogFormat      logger.LogFormat
}

func Default() Config {
	home, _ := os.UserHomeDir()

	return Config{
		RootDir:        filepath.Join(home, ".crun"),
		AppLogDir:      filepath.Join(os.TempDir(), "crun"),
		ConfigFilePath: filepath.Join(home, ".crun", "config.toml"),
		LogLevel:       logger.LevelInfo,
		LogFormat:      logger.JSONLogFormat,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if path == "" {
		path = cfg.ConfigFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
