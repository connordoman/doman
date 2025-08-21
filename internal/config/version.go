package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	Version = "0.0.1"
	Build   = "dev"

	// Dynamic

	CommitHash = "unknown"
	BuildDate  = "unknown"
)

type VersionSegment = int

const (
	BumpMajor VersionSegment = iota
	BumpMinor
	BumpPatch
)

type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

type VersionFile struct {
	Path    string
	Version VersionInfo
}

func OpenVersionFile(location string) (*VersionFile, error) {
	file, err := os.ReadFile(location)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		if err := os.MkdirAll(filepath.Dir(location), 0755); err != nil {
			return nil, err
		}

		_, err := os.Create(location)
		if err != nil {
			return nil, err
		}

		file = []byte("0.0.0")
	}

	var version = VersionInfo{}

	if len(file) == 0 {
		return nil, fmt.Errorf("version file is empty: %s", location)
	}

	if len(file) < 5 {
		return nil, fmt.Errorf("version file is not formatted correctly: %s", location)
	}

	if len(file) > 32 {
		return nil, fmt.Errorf("version file is too long: %s", location)
	}

	fileString := string(file)
	versionSegments := strings.Split(fileString, ".")

	if len(versionSegments) != 3 {
		return nil, fmt.Errorf("invalid version format in file %s: expected 'major.minor.patch', got '%s'", location, fileString)
	}

	version.Major, err = strconv.Atoi(versionSegments[0])
	if err != nil {
		return nil, err
	}
	version.Minor, err = strconv.Atoi(versionSegments[1])
	if err != nil {
		return nil, err
	}
	version.Patch, err = strconv.Atoi(versionSegments[2])
	if err != nil {
		return nil, err
	}

	return &VersionFile{
		Path:    location,
		Version: version,
	}, nil
}

func (vf VersionFile) String() string {
	return fmt.Sprintf("%d.%d.%d", vf.Version.Major, vf.Version.Minor, vf.Version.Patch)
}

func (vf *VersionFile) Save() error {
	return os.WriteFile(vf.Path, []byte(vf.String()), 0644)
}

func (vf VersionFile) Bump(segment VersionSegment) error {
	switch segment {
	case BumpMajor:
		vf.Version.Major++
		vf.Version.Minor = 0
		vf.Version.Patch = 0
	case BumpMinor:
		vf.Version.Minor++
		vf.Version.Patch = 0
	case BumpPatch:
		vf.Version.Patch++
	}
	return vf.Save()
}
