package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type LogLevel string

const ConfigFilePath string = "/etc/crun/config.toml"

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
	RootDir   string
	AppLogDir string
	LogFormat LogFormat
	LogLevel  LogLevel
}

func Default() Config {
	return Config{
		RootDir:   "/var/lib/crun",
		LogLevel:  LevelInfo,
		LogFormat: JSONLogFormat,
		AppLogDir: "/run/crun",
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if path == "" {
		path = ConfigFilePath
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
