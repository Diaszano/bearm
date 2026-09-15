package config

import (
	"bytes"
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

	return ApplyEnvironment(value, getenv)
}
