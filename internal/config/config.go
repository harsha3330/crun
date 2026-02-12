package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	RootDir        string
	AppLogDir      string
	ConfigFilePath string
}

func Default() Config {
	home, _ := os.UserHomeDir()

	return Config{
		RootDir:        filepath.Join(home, ".crun"),
		AppLogDir:      filepath.Join(os.TempDir(), "crun"),
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
