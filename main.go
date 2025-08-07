package main

import (
	"log"

	"github.com/connordoman/doman/cmd"
	"github.com/connordoman/doman/internal/txt"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(txt.Errorf("fatal: %v", err))
	}
}
