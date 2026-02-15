package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

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

func startContainerSimple(rootfs string, args []string, env []string, logFile *os.File, detached bool, hostNetwork bool) (int, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("no command specified")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Dir = "/"
	cmd.Stdin = os.Stdin

	if logFile != nil {
		if detached {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		} else {
			cmd.Stdout = io.MultiWriter(os.Stdout, logFile)
			cmd.Stderr = io.MultiWriter(os.Stderr, logFile)
		}
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cloneFlags := uintptr(0)
	if !hostNetwork {
		cloneFlags = syscall.CLONE_NEWNET
	}
	attr := &syscall.SysProcAttr{
		Chroot:     rootfs,
		Setpgid:    true,
		Cloneflags: cloneFlags,
	}
	if !detached {
		attr.Pdeathsig = syscall.SIGKILL
	}
	cmd.SysProcAttr = attr

	if err := cmd.Start(); err != nil {
		return 0, err
	}

	if detached {
		return cmd.Process.Pid, nil
	}

	pgid := -cmd.Process.Pid
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			_ = syscall.Kill(pgid, sig.(syscall.Signal))
		}
	}()

	err := cmd.Wait()
	signal.Stop(sigCh)
	close(sigCh)
	if err != nil {
		return cmd.Process.Pid, err
	}
	return cmd.Process.Pid, nil
}

type RunOptions struct {
	HostNetwork bool
}

func Run(cfg config.Config, log *slog.Logger, stater logger.Console, image string, opts *RunOptions) error {
	if opts == nil {
		opts = &RunOptions{}
	}
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

	mergedPath := filepath.Join(cfg.RootDir, "containers", containerId, "merged")
	if err := pkg.SetupDev(mergedPath); err != nil {
		stater.Error("failed to setup /dev", "error", err)
		return err
	}

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

	containerDir := filepath.Join(cfg.RootDir, "containers", containerId)
	logPath := filepath.Join(containerDir, "log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		stater.Error("failed to create container log file", "error", err)
		return err
	}
	defer logFile.Close()

	startedLine := fmt.Sprintf("=== Container started at %s | container-id: %s ===\n",
		time.Now().Format(time.RFC3339), containerId)
	if _, err := logFile.WriteString(startedLine); err != nil {
		stater.Warn("failed to write started line to log file", "error", err)
	}
	stater.Success("Container started",
		"container-id", containerId,
		"logs", logPath,
	)

	stater.Step("starting container process",
		"id", containerId,
		"rootfs", mergedPath,
		"cmd", processArgs,
	)
	pid, err := startContainerSimple(mergedPath, processArgs, configData.Config.Env, logFile, true, opts.HostNetwork)
	if err != nil {
		stater.Error("failed to start container process", "error", err)
		return err
	}

	pidPath := filepath.Join(containerDir, "pid")
	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		stater.Error("failed to write pid file", "error", err)
		_ = syscall.Kill(pid, syscall.SIGKILL)
		return err
	}

	stater.Success("container started (detached)",
		"container-id", containerId,
		"pid", pid,
		"logs", logPath,
	)
	log.Info("container started (detached)", "container-id", containerId, "pid", pid, "logs", logPath)
	stater.Step(fmt.Sprintf("stop: crun stop %s", containerId))
	stater.Step(fmt.Sprintf("logs: cat %s", logPath))
	if opts.HostNetwork {
		stater.Step("UI: http://localhost (e.g. port 80 for nginx)")
	} else {
		stater.Step("service: listening in container (e.g. port 80); use --network=host to access UI at http://localhost")
	}
	return nil
}
