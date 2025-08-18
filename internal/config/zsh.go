package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	zshrcContent string = ""
	zshrcLoaded  bool   = false
)

func IsUsingZsh() bool {
	shell := CheckShell()
	return strings.HasSuffix(shell, "/zsh") || strings.HasSuffix(shell, "\\zsh")
}

func ZshrcPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".zshrc")
}

func LoadZshrc() error {
	zshrcPath := ZshrcPath()
	if zshrcPath == "" {
		return fmt.Errorf("could not determine zshrc path")
	}

	content, err := os.ReadFile(zshrcPath)
	if err != nil {
		return fmt.Errorf("failed to read .zshrc file: %w", err)
	}

	zshrcContent = string(content)
	zshrcLoaded = true

	return nil
}

func GetZshrcContent() (string, error) {
	if !zshrcLoaded {
		if err := LoadZshrc(); err != nil {
			return "", fmt.Errorf("failed to load zshrc file: %w", err)
		}
	}
	return zshrcContent, nil
}

func ZshContains(s string) bool {
	zshrcContent, err := GetZshrcContent()
	if err != nil {
		return false
	}
	return strings.Contains(zshrcContent, s)
}

func AppendToZshrc(s string) error {
	zshrcContent, err := GetZshrcContent()
	if err != nil {
		return fmt.Errorf("failed to get zshrc content: %w", err)
	}

	zshrcContent += s

	if err := os.WriteFile(ZshrcPath(), []byte(zshrcContent), os.ModeAppend); err != nil {
		return fmt.Errorf("failed to save zshrc content: %w", err)
	}

	return nil
}
