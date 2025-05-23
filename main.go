package main

import (
	"log"

	"github.com/connordoman/doman/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
