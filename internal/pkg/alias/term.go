package alias

import (
	"fmt"

	"doman.sh/doman/internal/config"
	"doman.sh/doman/internal/pkg"
	"doman.sh/doman/internal/txt"
)

func PrintReloadWarning() {
	zshrcPath := config.ZshrcPath()
	if zshrcPath == "" {
		zshrcPath = "~/.zshrc"
	}

	sourceCommand := fmt.Sprintf("source %s", zshrcPath)

	fmt.Printf("! Be sure to run %s to apply the changes", txt.Boldf("%s", sourceCommand))
	err := pkg.CopyToClipboard(sourceCommand)
	if err != nil {
		fmt.Printf(" (not copied)\n")
	} else {
		fmt.Printf(" (copied to clipboard)\n")
	}
}
