package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
	"github.com/harsha3330/crun/internal/pkg"
)

func generateContainerID() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func newContainerID(root string) (string, error) {
	for {
		id, err := generateContainerID()
		if err != nil {
			return "", err
		}

		path := filepath.Join(root, "containers", id)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return id, nil
		}
	}
}

func constructLowerDir(cfg config.Config, layers []pkg.Descriptor) string {
	var lowers []string
	for i := len(layers) - 1; i >= 0; i-- {
		digest := layers[i].Digest[7:]
		path := filepath.Join(cfg.RootDir, "layers", digest)
		lowers = append(lowers, path)
	}
	return strings.Join(lowers, ":")
}

func createContainerDirs(cfg config.Config, containerId string, lowerdir string) error {
	containerDir := filepath.Join(cfg.RootDir, "containers", containerId)
	upper := filepath.Join(containerDir, "upper")
	work := filepath.Join(containerDir, "work")
	merged := filepath.Join(containerDir, "merged")

	if err := pkg.EnsureDir(upper); err != nil {
		return err
	}
	if err := pkg.EnsureDir(work); err != nil {
		return err
	}
	if err := pkg.EnsureDir(merged); err != nil {
		return err
	}
	options := fmt.Sprintf(
		"lowerdir=%s,upperdir=%s,workdir=%s",
		lowerdir, upper, work,
	)
	if err := syscall.Mount("overlay", merged, "overlay", 0, options); err != nil {
		return err
	}
	return nil
}

func buildProcessArgs(cfg pkg.OCIImageConfig) []string {
	entry := cfg.Config.Entrypoint
	cmd := cfg.Config.Cmd

	if len(entry) > 0 {
		return append(entry, cmd...)
	}
	return cmd
}

func Run(cfg config.Config, log *slog.Logger, stater logger.Console, image string) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("crun must be run as root for creating merged fs and mounting")
	}
	log.Info("Starting the process for the image", "value", image)
	stater.Step("Starting the image", "value", image)
	repo, tag, err := parseImage(image)
	stater.Step("Image Arguments", "repository", repo, "tag", tag)
	if err != nil {
		log.Error(err.Error())
		stater.Error(err.Error())
		return err
	}
	manifestDigestLocation := filepath.Join(cfg.RootDir, "images", repo, "tags", tag)
	digestData, err := os.ReadFile(manifestDigestLocation)
	if err != nil {
		stater.Error("error while getting the manifest digest for the image tag", "error", err)
		return err
	}
	digest := string(digestData)
	log.Debug("got the digest for image", "repo", repo, "tag", tag, "digest", digest)
	manifestLocation := filepath.Join(cfg.RootDir, "images", repo, "manifests", digest[7:], "manifest.json")
	manifestData, err := os.ReadFile(manifestLocation)
	if err != nil {
		stater.Error("error while getting the manifests data from manifest file", "location", manifestLocation)
	}
	stater.Success("got the manifests data from manifest file")
	var ociImageManifest pkg.OCIManifest
	err = json.Unmarshal(manifestData, &ociImageManifest)
	if err != nil {
		stater.Error("error decoding the image manifests", "error", err)
		return err
	}
	log.Debug("image manifets data for run command ", "value", ociImageManifest)

	containerId, err := newContainerID(cfg.RootDir)
	if err != nil {
		stater.Error("error generating new containerid", "error", err)
		return err
	}
	stater.Step("creating filssystem for container", "id", containerId)
	lowerDir := constructLowerDir(cfg, ociImageManifest.Layers)
	log.Info("constructed lower dir", "value", lowerDir)
	err = createContainerDirs(cfg, containerId, lowerDir)
	if err != nil {
		stater.Error("error creating container dirs / union filesystem", "error", err)
		return err
	}
	stater.Success("creating the merged filesystem ", "containerID", containerId)

	configDigest := ociImageManifest.Config.Digest[7:]
	configFilePath := filepath.Join(cfg.RootDir, "blobs", configDigest)
	configByteData, err := os.ReadFile(configFilePath)
	if err != nil {
		stater.Error("error reading config file of the image", "digest", configDigest)
		return err
	}
	stater.Success("succesfully read the config file of the image", "digest", configDigest)
	var configData pkg.OCIImageConfig
	err = json.Unmarshal(configByteData, &configData)
	if err != nil {
		stater.Error("error decoding config data", "error", err.Error())
		return err
	}
	stater.Success("decoded the config data to get runtime commands")
	log.Debug("config file", "data", configData)
	log.Debug("image execution config",
		"entrypoint", configData.Config.Entrypoint,
		"cmd", configData.Config.Cmd,
		"env_count", len(configData.Config.Env),
		"workdir", configData.Config.WorkingDir,
	)

	processArgs := buildProcessArgs(configData)
	if len(processArgs) == 0 {
		return fmt.Errorf("no command specified in image config")
	}

	mergedPath := filepath.Join(cfg.RootDir, "containers", containerId, "merged")
	stater.Step("starting container process",
		"id", containerId,
		"rootfs", mergedPath,
		"cmd", processArgs,
	)

	return nil
}
