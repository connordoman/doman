package main

import (
	"log"

	"doman.sh/cmd"
	"doman.sh/internal/txt"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(txt.Errorf("fatal: %v", err))
	}
}
