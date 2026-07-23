// Package platform contains operating-system-specific path and filesystem helpers.
package platform

import (
	"errors"
	"path/filepath"
)

// Dirs contains Bearm configuration, state, and data roots.
type Dirs struct {
	ConfigRoot string
	StateRoot  string
	DataRoot   string
}

// ResolveDirs resolves Bearm directories without reading global process state directly.
func ResolveDirs(getenv func(string) string, home, goos string) (Dirs, error) {
	if home == "" || !filepath.IsAbs(home) {
		return Dirs{}, errors.New("home directory must be absolute")
	}

	base := Dirs{}
	switch goos {
	case "darwin":
		support := filepath.Join(home, "Library", "Application Support", "Bearm")
		base = Dirs{
			ConfigRoot: support,
			StateRoot:  filepath.Join(support, "state"),
			DataRoot:   filepath.Join(support, "data"),
		}
	default:
		configHome := absoluteOrFallback(getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config"))
		stateHome := absoluteOrFallback(getenv("XDG_STATE_HOME"), filepath.Join(home, ".local", "state"))
		dataHome := absoluteOrFallback(getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share"))
		base = Dirs{
			ConfigRoot: filepath.Join(configHome, "bearm"),
			StateRoot:  filepath.Join(stateHome, "bearm"),
			DataRoot:   filepath.Join(dataHome, "bearm"),
		}
	}

	var err error
	if base.ConfigRoot, err = applyAbsoluteOverride(getenv("BEARM_CONFIG_HOME"), base.ConfigRoot); err != nil {
		return Dirs{}, err
	}
	if base.StateRoot, err = applyAbsoluteOverride(getenv("BEARM_STATE_HOME"), base.StateRoot); err != nil {
		return Dirs{}, err
	}
	if base.DataRoot, err = applyAbsoluteOverride(getenv("BEARM_DATA_HOME"), base.DataRoot); err != nil {
		return Dirs{}, err
	}

	return base, nil
}

func absoluteOrFallback(value, fallback string) string {
	if filepath.IsAbs(value) {
		return filepath.Clean(value)
	}
	return fallback
}

func applyAbsoluteOverride(value, fallback string) (string, error) {
	if value == "" {
		return fallback, nil
	}
	if !filepath.IsAbs(value) {
		return "", errors.New("Bearm directory overrides must be absolute")
	}
	return filepath.Clean(value), nil
}
