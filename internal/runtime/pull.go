package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
	"github.com/harsha3330/crun/internal/pkg"
)

const REGISTRY = "https://registry-1.docker.io"

func parseImage(image string) (string, string, error) {
	if image == "" {
		return "", "", fmt.Errorf("image argument is empty")
	}
	parts := strings.Split(image, ":")
	if len(parts) > 2 {
		return "", "", fmt.Errorf("unrecognized format image: %s ", image)
	}
	if len(parts) == 1 {
		return "", "", fmt.Errorf("empty tag recieved")
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("empty image repo or image recieved")
	}
	if parts[1] == "latest" {
		return "", "", fmt.Errorf("latest tag is not supported yet")
	}
	return parts[0], parts[1], nil
}

func normalizeRepo(repo string) string {
	if !strings.Contains(repo, "/") {
		return "library/" + repo
	}
	return repo
}

func getToken(repo string) (string, error) {
	repo = normalizeRepo(repo)
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Token, nil
}

func getImageIndex(repo, tag, token string) ([]byte, error) {
	url := fmt.Sprintf("%s/v2/library/%s/manifests/%s", REGISTRY, repo, tag)
	req, _ := http.NewRequest("GET", url, nil)
	bearerToken := "Bearer " + token
	req.Header.Add("Authorization", bearerToken)
	req.Header.Set("Accept",
		"application/vnd.oci.image.index.v1+json, "+
			"application/vnd.docker.distribution.manifest.list.v2+json, "+
			"application/vnd.oci.image.manifest.v1+json, "+
			"application/vnd.docker.distribution.manifest.v2+json",
	)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func getImageManifest(repo, digest, token string) ([]byte, error) {
	url := fmt.Sprintf("%s/v2/library/%s/blobs/%s", REGISTRY, repo, digest)
	req, _ := http.NewRequest("GET", url, nil)
	bearerToken := "Bearer " + token
	req.Header.Add("Authorization", bearerToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func DownloadBlob(repo, digest, token, destDir string) error {
	url := fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", repo, digest)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s: %s", digest, resp.Status)
	}
	filename := filepath.Join(destDir, digest[7:])
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func DownloadImageBlobs(repo string, config pkg.Descriptor, layers []pkg.Descriptor, destDir string, log *slog.Logger, stater logger.Console) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	token, err := getToken(repo)
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	errCh := make(chan error, len(layers))
	download := func(digest string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		if err := DownloadBlob(repo, digest, token, destDir); err != nil {
			log.Error("error downloading %s: %v", digest, err)
			stater.Error("error downloading blob", "digest", digest, "error", err.Error())
			errCh <- err
			return
		} else {
			log.Info("downloaded blob", "digest", digest)
			stater.Success("downloaded blob", "digest", digest)
		}
	}
	wg.Add(1)
	go download(config.Digest)
	for _, layer := range layers {
		wg.Add(1)
		go download(layer.Digest)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

func extractImage(blobDir, layerDir string, layers []pkg.Descriptor, log *slog.Logger, stater logger.Console) error {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 4)
	errCh := make(chan error, len(layers))
	extract := func(blobDir, layerDir, digest string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()

		fspath, err := pkg.EnsureLayerExtracted(blobDir, layerDir, digest)
		if err != nil {
			log.Error("error extracting image layer", "digest", digest)
			stater.Error("error extracting image layer", "digest", digest)
			errCh <- err
			return
		} else {
			log.Info("extracted image layer", "digest", digest, "filepath", fspath)
			stater.Success("succesfully extracted image layer into fs", "digest", digest, "filepath", fspath)
		}
	}

	for _, layer := range layers {
		digest := strings.TrimPrefix(layer.Digest, "sha256:")
		wg.Add(1)
		go extract(blobDir, layerDir, digest)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		return err
	}
	return nil
}

func Pull(cfg config.Config, log *slog.Logger, stater logger.Console, image string) error {
	log.Info("Starting pull the image", "value", image)
	stater.Step("Pulling the image", "value", image)
	repo, tag, err := parseImage(image)
	log.Debug("recived the following image contents", "image repo", repo, "image tag", tag)
	stater.Step("Image Arguments", "repository", repo, "tag", tag)
	if err != nil {
		log.Error(err.Error())
		stater.Error(err.Error())
		return err
	}
	stater.Step("Getting auth to pull image manifests")
	token, err := getToken(repo)
	if err != nil {
		stater.Error("Failed to get the bearer token of the repository", "repository", repo)
		return err
	}
	stater.Success("Got the bearer token of the repository", "repository", repo)
	stater.Step("Getting the image index")
	imageIndexData, err := getImageIndex(repo, tag, token)
	if err != nil {
		stater.Error("error getting the image index data", "repo", repo, "tag", tag)
		return err
	}
	stater.Success("got the image index data", "repo", repo, "tag", tag)
	log.Debug("image manifest file", "content", imageIndexData)
	stater.Step("Decoding image index data")
	ociIndex, err := pkg.DecodeIndex(imageIndexData)
	if err != nil {
		stater.Error("error decoding the image index data")
		return err
	}
	stater.Success("Decoded the image index data")
	log.Debug("Index Decoded content", "OCI Index", ociIndex)
	platform := pkg.HostPlatform()
	stater.Step("Got the platform details", "OS", platform.OS, "Architecture", platform.Arch)
	imageDigest, err := pkg.SelectPlatformManifest(ociIndex, platform.OS, platform.Arch)
	if err != nil {
		stater.Error("error getting the manifest for this platform", "os", platform.OS, "architecture", platform.Arch)
		return err
	}
	log.Debug("Platform Manifest Digest", "Manifest", imageDigest, "OS", platform.OS, "Architecture", platform.Arch)
	stater.Step("Getting Image Layers , Configs")
	ociManifestData, err := getImageManifest(repo, imageDigest, token)
	if err != nil {
		stater.Error("Error getting image manifests (contains config , layers)")
		return err
	}
	log.Debug("Image Manifest", "Data", ociManifestData)
	imageManifest, err := pkg.DecodeManifestAuto(ociManifestData)
	if err != nil {
		stater.Error("Error decoding image manifests (contains config , layers)")
		return err
	}

	err = pkg.SaveFile(filepath.Join(cfg.RootDir, "images", repo, "manifests", imageDigest[7:], "manifest.json"), ociManifestData)
	if err != nil {
		stater.Error("error saving the manifests file")
		return err
	}
	stater.Success("saved the manifest file")
	log.Debug("Decoded Image Manifest", "value", imageManifest)
	stater.Success("Got the image manifest data , layers , config")
	config, layers := imageManifest.Config, imageManifest.Layers
	err = DownloadImageBlobs(normalizeRepo(repo), config, layers, filepath.Join(cfg.RootDir, "blobs"), log, stater)
	if err != nil {
		stater.Error("Error Downloading image blobs")
		return err
	}
	tagFilePath := filepath.Join(cfg.RootDir, "images", repo, "tags", tag)
	err = pkg.SaveFile(tagFilePath, []byte(imageDigest))
	if err != nil {
		stater.Error("error saving tag file data of the manifests")
	}

	blobDir, layerDir := filepath.Join(cfg.RootDir, "blobs"), filepath.Join(cfg.RootDir, "layers")
	err = extractImage(blobDir, layerDir, layers, log, stater)
	if err != nil {
		stater.Error("error extracting layers into filesystem", "error", err.Error())
		return err
	}
	stater.Success("extracted all the layers into filesystem , image pull completed")
	return nil
}
