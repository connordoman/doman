package pkg

import "github.com/connordoman/doman/internal/config"

func VersionString() string {
	return "doman " + config.Version + " (#" + config.CommitHash[:7] + ") " + config.Build + " " + config.BuildDate
}
