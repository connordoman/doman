package pkg

import (
	"fmt"
	"os"

	"github.com/connordoman/doman/internal/txt"
)

func PrintSuccess(format string, args ...any) {
	str := fmt.Sprintf("✔ "+format, args...)
	fmt.Println(txt.Successf(str))
}

func PrintError(format string, args ...any) {
	str := fmt.Sprintf("✘ "+format, args...)
	fmt.Fprintln(os.Stderr, txt.Errorf(str))
}

func PrintInfo(format string, args ...any) {
	fmt.Println(txt.Greyf(format, args...))
}

func FailAndExit(format string, args ...any) {
	PrintError(format, args...)
	os.Exit(1)
}
