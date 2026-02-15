package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harsha3330/crun/internal/config"
	logger "github.com/harsha3330/crun/internal/log"
	"github.com/harsha3330/crun/internal/pkg"
)

// RemoveImage removes a pulled image: deletes its tag and manifest, then removes
// blobs and extracted layers that are not referenced by any other image.
// It fails if any container is still using this image.
func RemoveImage(cfg config.Config, stater logger.Console, image string) error {
	repo, tag, err := parseImage(image)
	if err != nil {
		stater.Error("invalid image", "error", err)
		return err
	}

	// Refuse to remove if any container is using this image
	if ids := containersUsingImage(cfg.RootDir, image); len(ids) > 0 {
		stater.Error("image in use by container(s)", "image", image, "container-id", ids[0])
		return fmt.Errorf("image %s is in use by container %s (stop it first)", image, ids[0])
	}

	tagPath := filepath.Join(cfg.RootDir, "images", repo, "tags", tag)
	digestData, err := os.ReadFile(tagPath)
	if err != nil {
		if os.IsNotExist(err) {
			stater.Error("image not found", "image", image)
			return fmt.Errorf("image not found: %s", image)
		}
		stater.Error("failed to read tag file", "error", err)
		return err
	}
	digest := string(digestData)
	digestTrimmed := digest
	if len(digestTrimmed) > 7 {
		digestTrimmed = digestTrimmed[7:] // strip "sha256:"
	}

	manifestDir := filepath.Join(cfg.RootDir, "images", repo, "manifests", digestTrimmed)
	manifestPath := filepath.Join(manifestDir, "manifest.json")

	// Read manifest before deleting so we can remove blobs/layers not used by other images
	var digestsToMaybeRemove []string
	if data, err := os.ReadFile(manifestPath); err == nil {
		var m pkg.OCIManifest
		if json.Unmarshal(data, &m) == nil {
			if len(m.Config.Digest) > 7 {
				digestsToMaybeRemove = append(digestsToMaybeRemove, m.Config.Digest[7:])
			}
			for _, l := range m.Layers {
				if len(l.Digest) > 7 {
					digestsToMaybeRemove = append(digestsToMaybeRemove, l.Digest[7:])
				}
			}
		}
	}

	if err := os.Remove(tagPath); err != nil {
		stater.Error("failed to remove tag file", "error", err)
		return err
	}
	if err := os.RemoveAll(manifestDir); err != nil {
		stater.Error("failed to remove manifest dir", "error", err)
		return err
	}

	// Build set of digests still referenced by any remaining image
	inUse := referencedDigests(cfg.RootDir)

	blobDir := filepath.Join(cfg.RootDir, "blobs")
	layerDir := filepath.Join(cfg.RootDir, "layers")
	for _, d := range digestsToMaybeRemove {
		if inUse[d] {
			continue
		}
		_ = os.Remove(filepath.Join(blobDir, d))
		_ = os.RemoveAll(filepath.Join(layerDir, d))
	}

	cleanEmptyParents(cfg.RootDir, repo)
	stater.Success("image removed", "image", image)
	return nil
}

func containersUsingImage(rootDir, image string) []string {
	var ids []string
	containersDir := filepath.Join(rootDir, "containers")
	entries, _ := os.ReadDir(containersDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		imagePath := filepath.Join(containersDir, e.Name(), "image")
		data, err := os.ReadFile(imagePath)
		if err != nil {
			continue
		}
		if string(data) == image {
			ids = append(ids, e.Name())
		}
	}
	return ids
}

// referencedDigests returns a set (map) of all blob/layer digest suffixes (without "sha256:")
// that are still referenced by any tag in images/.
func referencedDigests(rootDir string) map[string]bool {
	out := make(map[string]bool)
	imagesDir := filepath.Join(rootDir, "images")
	entries, _ := os.ReadDir(imagesDir)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		tagsDir := filepath.Join(imagesDir, e.Name(), "tags")
		tagEntries, _ := os.ReadDir(tagsDir)
		for _, te := range tagEntries {
			if te.IsDir() {
				continue
			}
			tagPath := filepath.Join(tagsDir, te.Name())
			data, err := os.ReadFile(tagPath)
			if err != nil {
				continue
			}
			digest := string(data)
			trimmed := digest
			if len(trimmed) > 7 {
				trimmed = trimmed[7:]
			}
			manifestPath := filepath.Join(rootDir, "images", e.Name(), "manifests", trimmed, "manifest.json")
			manifestData, err := os.ReadFile(manifestPath)
			if err != nil {
				continue
			}
			var m pkg.OCIManifest
			if json.Unmarshal(manifestData, &m) != nil {
				continue
			}
			if len(m.Config.Digest) > 7 {
				out[m.Config.Digest[7:]] = true
			}
			for _, l := range m.Layers {
				if len(l.Digest) > 7 {
					out[l.Digest[7:]] = true
				}
			}
		}
	}
	return out
}

func cleanEmptyParents(rootDir, repo string) {
	repoDir := filepath.Join(rootDir, "images", repo)
	tagsDir := filepath.Join(repoDir, "tags")
	manifestsDir := filepath.Join(repoDir, "manifests")
	_ = os.Remove(tagsDir)       
	_ = os.Remove(manifestsDir)  
	_ = os.Remove(repoDir)      
}
