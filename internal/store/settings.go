package store

import (
	"os"
	"path/filepath"

	"orchard/internal/gitx"
)

// DefaultSettings are the preferences of a fresh install.
func DefaultSettings() Settings {
	return Settings{
		Theme:            "dark",
		Concurrency:      4,
		CloneDir:         defaultCloneDir(),
		PullStrategy:     gitx.PullFastForward,
		AutoStash:        true,
		AutoFetchMinutes: 15,
		ScanDepth:        4,
	}
}

func (s Settings) withDefaults() Settings {
	d := DefaultSettings()
	if s.Theme == "" {
		s.Theme = d.Theme
	}
	if s.Concurrency <= 0 {
		s.Concurrency = d.Concurrency
	}
	if s.CloneDir == "" {
		s.CloneDir = d.CloneDir
	}
	if s.PullStrategy == "" {
		s.PullStrategy = d.PullStrategy
	}
	if s.ScanDepth <= 0 {
		s.ScanDepth = d.ScanDepth
	}
	return s
}

func defaultCloneDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, "Code")
}
