package pkg

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func CheckPath(path string, isDir bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return err
	}

	if isDir && !info.IsDir() {
		return fmt.Errorf("%s exists but is not a directory", path)
	}

	if !isDir && info.IsDir() {
		return fmt.Errorf("%s is a directory but a file was expected", path)
	}

	return nil
}

func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path, 0777)
}

func EnsureFile(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("%s exists but is a directory", path)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0766); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}

func EnsureLayerExtracted(blobDir, layerDir, digest string) (string, error) {
	fsPath := filepath.Join(layerDir, digest)
	if _, err := os.Stat(fsPath); err == nil {
		return fsPath, nil
	}
	blobPath := filepath.Join(blobDir, digest)
	if err := os.MkdirAll(fsPath, 0755); err != nil {
		return "", err
	}
	if err := extractTarGz(blobPath, fsPath); err != nil {
		return "", err
	}
	return fsPath, nil
}

func extractTarGz(tarPath, dest string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
			if err := os.Chmod(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil && !os.IsExist(err) {
				return err
			}
		}
	}

	return nil
}

func mkdev(major, minor int) int {
	return (major << 8) | minor
}

func SetupDev(rootfs string) error {
	dev := filepath.Join(rootfs, "dev")
	if err := os.RemoveAll(dev); err != nil {
		return err
	}
	if err := os.MkdirAll(dev, 0755); err != nil {
		return err
	}
	if err := syscall.Mknod(
		filepath.Join(dev, "null"),
		syscall.S_IFCHR|0666,
		mkdev(1, 3),
	); err != nil {
		return err
	}
	if err := syscall.Mknod(
		filepath.Join(dev, "zero"),
		syscall.S_IFCHR|0666,
		mkdev(1, 5),
	); err != nil {
		return err
	}
	if err := syscall.Mknod(
		filepath.Join(dev, "random"),
		syscall.S_IFCHR|0666,
		mkdev(1, 8),
	); err != nil {
		return err
	}
	if err := syscall.Mknod(
		filepath.Join(dev, "urandom"),
		syscall.S_IFCHR|0666,
		mkdev(1, 9),
	); err != nil {
		return err
	}

	return nil
}
