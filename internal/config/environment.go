package config

import (
	"fmt"
	"os"
	"strconv"
)

// ApplyEnvironment applies exact Bearm environment overrides.
func ApplyEnvironment(value Config, getenv func(string) string) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	if current := getenv("BEARM_LANG"); current != "" {
		value.Language = current
	}
	if current := getenv("BEARM_COMPAT"); current != "" {
		value.CompatibilityProfile = current
	}
	if current := getenv("BEARM_TRASH"); current != "" {
		value.Trash.CustomPath = current
	}
	if current := getenv("BEARM_LOG"); current != "" {
		value.Logging.Level = current
	}

	var err error
	if current := getenv("BEARM_TRASH_PER_MOUNT"); current != "" {
		value.Trash.PerMount, err = strconv.ParseBool(current)
		if err != nil {
			return Config{}, fmt.Errorf("BEARM_TRASH_PER_MOUNT: %w", err)
		}
	}

	if err := value.Validate(); err != nil {
		return Config{}, err
	}
	return value, nil
}
