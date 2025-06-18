package pkg

import "github.com/connordoman/doman/internal/config"

func Version() string {
	return config.Version + " (" + config.CommitHash[:7] + ") " + config.Build + " " + config.BuildDate
}
