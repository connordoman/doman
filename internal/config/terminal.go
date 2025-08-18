package config

import "os"

func CheckShell() string {
	return os.Getenv("SHELL")
}
