// Package config loads and validates Bearm configuration.
package config

import (
	"errors"
	"fmt"
	"path/filepath"
)

// Config is the complete Bearm configuration.
type Config struct {
	Language             string        `toml:"language"`
	CompatibilityProfile string        `toml:"compatibility_profile"`
	Trash                TrashConfig   `toml:"trash"`
	Safety               SafetyConfig  `toml:"safety"`
	Restore              RestoreConfig `toml:"restore"`
	Logging              LoggingConfig `toml:"logging"`
}

// TrashConfig controls platform trash selection.
type TrashConfig struct {
	PerMount   bool   `toml:"per_mount"`
	CustomPath string `toml:"custom_path"`
}

// SafetyConfig controls root, scope, and pattern protection.
type SafetyConfig struct {
	PreserveRoot       bool     `toml:"preserve_root"`
	AllowedRoots       []string `toml:"allowed_roots"`
	ProtectedPatterns  []string `toml:"protected_patterns"`
	InspectDescendants bool     `toml:"inspect_descendants"`
}

// RestoreConfig controls restore collisions.
type RestoreConfig struct {
	CollisionPolicy string `toml:"collision_policy"`
}

// LoggingConfig controls local diagnostic verbosity.
type LoggingConfig struct {
	Level string `toml:"level"`
}

// Default returns exact version 1 platform-independent defaults.
func Default() Config {
	return Config{
		Language:             "pt-BR",
		CompatibilityProfile: "auto",
		Trash: TrashConfig{
			PerMount: true,
		},
		Safety: SafetyConfig{
			PreserveRoot:       true,
			AllowedRoots:       []string{},
			ProtectedPatterns:  []string{},
			InspectDescendants: false,
		},
		Restore: RestoreConfig{CollisionPolicy: "fail"},
		Logging: LoggingConfig{Level: "error"},
	}
}

// Validate validates all configuration values.
func (c Config) Validate() error {
	if c.Language != "pt-BR" && c.Language != "en" {
		return fmt.Errorf("unsupported language %q", c.Language)
	}
	switch c.CompatibilityProfile {
	case "auto", "gnu", "bsd", "posix":
	default:
		return fmt.Errorf("unsupported compatibility profile %q", c.CompatibilityProfile)
	}
	if c.Trash.CustomPath != "" && !filepath.IsAbs(c.Trash.CustomPath) {
		return errors.New("trash.custom_path must be absolute")
	}
	for _, root := range c.Safety.AllowedRoots {
		if !filepath.IsAbs(root) {
			return fmt.Errorf("safety.allowed_roots entry must be absolute: %q", root)
		}
	}
	switch c.Restore.CollisionPolicy {
	case "fail", "rename", "overwrite":
	default:
		return fmt.Errorf("unsupported restore collision policy %q", c.Restore.CollisionPolicy)
	}
	switch c.Logging.Level {
	case "error", "debug":
	default:
		return fmt.Errorf("unsupported logging level %q", c.Logging.Level)
	}
	return nil
}
