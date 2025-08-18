package alias

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/connordoman/doman/internal/config"
	"github.com/connordoman/doman/internal/pkg"
)

const (
	AliasFolderName   = "aliases"
	AliasLoaderScript = "aliases.zsh"
)

//go:embed aliases.zsh
var aliasesScriptContent string

func AliasFolderPath() (string, error) {
	configDir, err := config.GetConfigPath()
	if err != nil {
		return "", fmt.Errorf("could not find config folder: %w", err)
	}

	aliasPath := filepath.Join(configDir, AliasFolderName)
	return aliasPath, nil
}

func AliasFolderExists() bool {
	aliasPath, err := AliasFolderPath()
	if err != nil {
		return false
	}
	return pkg.DirExists(aliasPath)
}

func CreateAliasFolder() error {
	aliasPath, err := AliasFolderPath()
	if err != nil {
		return err
	}

	if !pkg.DirExists(aliasPath) {
		if err := pkg.Mkdir(aliasPath); err != nil {
			return err
		}
	}

	return nil
}

func AliasLoaderScriptPath() (string, error) {
	configDir, err := config.GetConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get config path: %w", err)
	}

	scriptPath := filepath.Join(configDir, AliasLoaderScript)

	return scriptPath, nil
}

func CreateAliasLoaderScript() error {
	if aliasesScriptContent == "" {
		return fmt.Errorf("aliases script content is empty")
	}

	aliasLoaderScriptPath, err := AliasLoaderScriptPath()
	if err != nil {
		return err
	}

	return pkg.WriteFile(aliasLoaderScriptPath, []byte(aliasesScriptContent))
}

func AddAliasLoaderScriptToZshrc() error {
	zshrcContent, err := config.GetZshrcContent()
	if err != nil {
		return fmt.Errorf("failed to get zshrc content: %w", err)
	}

	aliasLoaderScriptPath, err := AliasLoaderScriptPath()
	if err != nil {
		return err
	}
	additionalContent := fmt.Sprintf("\n# Loader script for zsh aliases added using `doman alias`\nsource \"%s\"\n", aliasLoaderScriptPath)

	if !strings.Contains(zshrcContent, additionalContent) {
		if err := config.AppendToZshrc(additionalContent); err != nil {
			return fmt.Errorf("failed to append to zshrc: %w", err)
		}
	}

	return nil
}
