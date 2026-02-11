package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

type LogFormat string

const (
	JSONLogFormat LogFormat = "json"
	TextLogFormat LogFormat = "text"
)

type Config struct {
	RootDir        string
	AppLogDir      string
	LogFormat      LogFormat
	LogLevel       LogLevel
	ConfigFilePath string
}

func Default() Config {
	home, _ := os.UserHomeDir()

	return Config{
		RootDir:        filepath.Join(home, ".crun"),
		AppLogDir:      filepath.Join(os.TempDir(), "crun"),
		LogLevel:       LevelInfo,
		LogFormat:      JSONLogFormat,
		ConfigFilePath: filepath.Join(home, ".crun", "config.toml"),
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
