package pkg

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type CommandConfig struct {
	EchoOff bool
}

var cmdConfig CommandConfig = CommandConfig{
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

func runCommandHelper(withOutput, printCommand bool, command string, args ...string) (string, error) {
	if printCommand {
		cmdString := fmt.Sprintf("$ %s %s", command, strings.Join(args, " "))
		log.Println(Bold(Gray(cmdString)))
	}

	cmd := exec.Command(command, args...)

	var outBuffer bytes.Buffer

	if withOutput {
		// For commands that should output directly to terminal to preserve colors
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// For commands where we need to capture output
		cmd.Stdout = &outBuffer
		cmd.Stderr = &outBuffer
	}

	if err := cmd.Run(); err != nil {
		return outBuffer.String(), fmt.Errorf("error running command '%s %s': %w", command, strings.Join(args, " "), err)
	}

	out := outBuffer.String()

	return out, nil
}

func RunCommand(command string, args ...string) (string, error) {
	return runCommandHelper(false, !cmdConfig.EchoOff, command, args...)
}

func RunCommandWithOutput(command string, args ...string) (string, error) {
	return runCommandHelper(true, !cmdConfig.EchoOff, command, args...)
}

func SetEchoOff() {
	cmdConfig.EchoOff = true
}

func SetEchoOn() {
	cmdConfig.EchoOff = false
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
