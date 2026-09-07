package alias

import (
	"fmt"

	"doman.sh/internal/config"
	"doman.sh/internal/pkg"
	"doman.sh/internal/txt"
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
