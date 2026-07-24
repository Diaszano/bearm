package config

import (
	"bytes"
	"errors"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Load loads defaults, strict TOML, environment overrides, and validation.
func Load(path string, getenv func(string) string) (Config, error) {
	value := Default()

	if err := CheckSecureFile(path, os.Getuid()); err != nil {
		if os.IsNotExist(err) {
			return ApplyEnvironment(value, getenv)
		}
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ApplyEnvironment(value, getenv)
		}
		return Config{}, err
	}

	decoder := toml.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return Config{}, err
	}

	value, err = ApplyEnvironment(value, getenv)
	if err != nil {
		return Config{}, err
	}
	return value, nil
}

// IsMissing reports whether a config error is a missing file.
func IsMissing(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}
