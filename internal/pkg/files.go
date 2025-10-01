package pkg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func Chdir(path string) error {
	if err := os.Chdir(path); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", path, err)
	}
	return nil
}

func Mkdir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func WriteFile(path string, content []byte) error {
	// if exists := FileExists(path); !exists {
	// 	if _, err := os.Create(path); err != nil {
	// 		return fmt.Errorf("failed to create file %s: %w", path, err)
	// 	}

	// }
	if err := os.WriteFile(path, content, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func ReadFile(path ...string) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("no file path provided")
	}

	content, err := os.ReadFile(filepath.Join(path...))
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filepath.Join(path...), err)
	}
	return content, nil
}
