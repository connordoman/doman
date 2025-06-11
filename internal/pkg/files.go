package pkg

import (
	"errors"
	"fmt"
	"os"
)

func Cwd() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	return dir, nil
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func Chdir(path string) error {
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", path, err)
	}
	return nil
}
