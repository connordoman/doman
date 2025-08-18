package alias

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/connordoman/doman/internal/pkg"
)

type Alias struct {
	Name        string
	Command     string
	Description string
}

func contains(s, token string) bool {
	return strings.Contains(s, token)
}

func NewAlias(name, command string) (*Alias, error) {
	if name == "" {
		return nil, fmt.Errorf("alias name cannot be empty")
	} else if command == "" {
		return nil, fmt.Errorf("alias command cannot be empty")
	}

	if contains(name, " ") {
		return nil, fmt.Errorf("alias name cannot contain spaces")
	}

	if contains(command, "\"") {
		return nil, fmt.Errorf("alias command cannot contain double quotes")
	}

	if contains(command, "\n") {
		return nil, fmt.Errorf("alias command cannot contain newlines")
	}

	return &Alias{
		Name:    name,
		Command: command,
	}, nil
}

func (a *Alias) String() string {
	var result string = ""

	if a.Description != "" {
		desc := "# " + strings.ReplaceAll(a.Description, "\n", "\n# ")
		result += fmt.Sprintf("%s\n", desc)
	}
	result += fmt.Sprintf("alias %s=\"%s\"", a.Name, a.Command)
	return result
}

func (a *Alias) Describe(desc string) {
	a.Description = desc
}

func (a Alias) Save() (string, error) {
	aliasesDir, err := AliasFolderPath()
	if err != nil {
		return "", err
	}

	aliasPath := filepath.Join(aliasesDir, a.Name+".zsh")

	if err := pkg.WriteFile(aliasPath, []byte(a.String())); err != nil {
		return "", err
	}

	return aliasPath, nil
}
