package pkg

import (
	"fmt"
	"os"
	"path/filepath"
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
