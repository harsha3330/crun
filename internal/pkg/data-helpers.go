package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type OCIIndex struct {
	SchemaVersion int    `json:"schemaVersion"`
	MediaType     string `json:"mediaType"`
	Manifests     []struct {
		MediaType string `json:"mediaType"`
		Digest    string `json:"digest"`
		Size      int64  `json:"size"`
		Platform  struct {
			Architecture string `json:"architecture"`
			OS           string `json:"os"`
			Variant      string `json:"variant,omitempty"`
		} `json:"platform"`
	} `json:"manifests"`
}

type Descriptor struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

type OCIManifest struct {
	SchemaVersion int          `json:"schemaVersion"`
	MediaType     string       `json:"mediaType"`
	Config        Descriptor   `json:"config"`
	Layers        []Descriptor `json:"layers"`
}

type OCIImageConfig struct {
	Architecture string `json:"architecture"`
	OS           string `json:"os"`

	Config struct {
		Env          []string       `json:"Env"`
		Entrypoint   []string       `json:"Entrypoint"`
		Cmd          []string       `json:"Cmd"`
		WorkingDir   string         `json:"WorkingDir"`
		ExposedPorts map[string]any `json:"ExposedPorts"`
		StopSignal   string         `json:"StopSignal"`
	} `json:"config"`
}

func DecodeImageManifest(data []byte) (*OCIManifest, error) {
	raw, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("base64 decode failed: %w", err)
	}

	var manifest OCIManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("json decode failed: %w", err)
	}

	return &manifest, nil
}

func DecodeManifestAuto(data []byte) (*OCIManifest, error) {
	var raw []byte

	if len(data) > 0 && data[0] == '{' {
		raw = data
	} else {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("data is neither JSON nor valid base64: %w", err)
		}
		raw = decoded
	}

	var manifest OCIManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	return &manifest, nil
}

func DecodeIndex(data []byte) (*OCIIndex, error) {
	if len(data) > 3 && string(data[:3]) == "eyJ" {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil {
			data = decoded
		}
	}

	var idx OCIIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, err
	}

	return &idx, nil
}

func SelectPlatformManifest(idx *OCIIndex, os, arch string) (string, error) {
	for _, m := range idx.Manifests {
		if m.Platform.OS == os && m.Platform.Architecture == arch {
			return m.Digest, nil
		}
	}
	return "", fmt.Errorf("no manifest for %s/%s", os, arch)
}

func SaveFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return os.Chmod(path, 0444)
}
