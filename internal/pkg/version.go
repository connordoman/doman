package pkg

import "doman.sh/internal/config"

func VersionString() string {
	commit := config.CommitHash
	if commit == "" || commit == "unknown" {
		commit = "unknown"
	} else if len(commit) > 7 {
		commit = commit[:7]
	}

	return "doman " + config.Version + " (#" + commit + ") " + config.Build + " " + config.BuildDate
}
