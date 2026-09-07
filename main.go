package main

import (
	"log"

	"doman.sh/doman/cmd"
	"doman.sh/doman/internal/txt"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(txt.Errorf("fatal: %v", err))
	}
}
