package pkg

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type CommandConfig struct {
	EchoOff bool
}

var config CommandConfig = CommandConfig{
	EchoOff: false,
}

var (
	Green     = color.New(color.FgGreen).SprintfFunc()
	Red       = color.New(color.FgRed).SprintfFunc()
	Bold      = color.New(color.Bold).SprintfFunc()
	Italic    = color.New(color.Italic).SprintfFunc()
	Underline = color.New(color.Underline).SprintfFunc()
	Magenta   = color.New(color.FgMagenta).SprintfFunc()
	Cyan      = color.New(color.FgCyan).SprintfFunc()
	White     = color.New(color.FgWhite).SprintfFunc()
	Black     = color.New(color.FgBlack).SprintfFunc()
	Yellow    = color.New(color.FgYellow).SprintfFunc()
	Blue      = color.New(color.FgBlue).SprintfFunc()
	Gray      = color.New(color.FgHiBlack).SprintfFunc()
)

func runCommandHelper(printCommand bool, command string, args ...string) (string, error) {
	if printCommand {
		cmdString := fmt.Sprintf("$ %s %s", command, strings.Join(args, " "))
		log.Println(Bold(Gray(cmdString)))
	}

	cmd := exec.Command(command, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func RunCommand(command string, args ...string) (string, error) {
	return runCommandHelper(!config.EchoOff, command, args...)
}

func SetEchoOff() {
	config.EchoOff = true
}

func SetEchoOn() {
	config.EchoOff = false
}

func SetEcho(cmd *cobra.Command) {
	echoOn, err := cmd.Flags().GetBool("echo")
	if err != nil {
		return
	}

	if echoOn {
		SetEchoOn()
	} else {
		SetEchoOff()
	}
}
