package config

// These values are set at build time via ldflags.
var (
	Version    = "dev"
	CommitHash = "unknown"
	BuildDate  = "unknown"
	Build      = "dev"
)
