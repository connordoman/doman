package go_self

import (
	"strings"

	"github.com/connordoman/doman/internal/pkg"
	"github.com/spf13/cobra"
)

var (
	RunCommand = &cobra.Command{
		Use:   "go",
		Short: "Run `go run main.go` slightly faster",
		Long:  "This command will run your Go code with some optimizations.",
		Run: func(cmd *cobra.Command, args []string) {
			pkg.RunCommandWithOutput("go", "run", "main.go", strings.Join(args, " "))
		},
	}
)
