package runtime

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	logger "github.com/harsha3330/crun/internal/log"
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
		return parts[0], "latest", nil
	}
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("empty image repo or image recieved")
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

func getImageManifest(repo, tag, token string) ([]byte, error) {
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

func Pull(log *slog.Logger, stater logger.Console, image string) error {
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
	imageManifestData, err := getImageManifest(repo, tag, token)
	log.Debug("image manifest file", "content", imageManifestData)
	return nil
}
