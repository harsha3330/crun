package pkg

import (
	"fmt"
	"os"
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

func EnsurePath(path string, isDir bool) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() && isDir {
			return fmt.Errorf("%s exists but is not a directory", path)
		}

		if info.IsDir() && !isDir {
			return fmt.Errorf("%s exists but is not a file", path)
		}
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(path, 0700)
}
