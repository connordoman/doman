package pkg

import (
	"fmt"

	"golang.design/x/clipboard"
)

var canCopy bool

func init() {
	err := clipboard.Init()
	if err != nil {
		canCopy = false
	} else {
		canCopy = true
	}
}

func CopyToClipboard(text string) error {
	if !canCopy {
		return fmt.Errorf("clipboard access is not available")
	}
	clipboard.Write(clipboard.FmtText, []byte(text))
	return nil
}
