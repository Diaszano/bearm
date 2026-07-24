package config_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestEnvironmentOverridesFileValues(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"BEARM_LANG":            "en",
		"BEARM_COMPAT":          "gnu",
		"BEARM_TRASH":           "/tmp/bearm-trash",
		"BEARM_TRASH_PER_MOUNT": "false",
		"BEARM_LOG":             "debug",
	}
	value := config.Default()
	got, err := config.ApplyEnvironment(value, func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("ApplyEnvironment() error = %v", err)
	}

	if got.Language != "en" ||
		got.CompatibilityProfile != "gnu" ||
		got.Trash.CustomPath != "/tmp/bearm-trash" ||
		got.Trash.PerMount ||
		got.Logging.Level != "debug" {
		t.Fatalf("ApplyEnvironment() = %#v", got)
	}
}

func TestEnvironmentRejectsInvalidBoolean(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_TRASH_PER_MOUNT": "sometimes"}
	if _, err := config.ApplyEnvironment(config.Default(), func(key string) string { return env[key] }); err == nil {
		t.Fatal("ApplyEnvironment() error = nil, want non-nil")
	}
}

func TestEnvironmentRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_LOG": "trace"}
	if _, err := config.ApplyEnvironment(config.Default(), func(key string) string { return env[key] }); err == nil {
		t.Fatal("ApplyEnvironment() error = nil, want non-nil for invalid log level")
	}
}

func TestEnvironmentHandlesNilGetenv(t *testing.T) {
	// Don't run in parallel as we modify real env or rely on fallback.
	// Since we pass nil, it will use os.Getenv. Let's make sure it doesn't panic.
	_, err := config.ApplyEnvironment(config.Default(), nil)
	if err != nil {
		t.Fatalf("ApplyEnvironment(..., nil) error = %v", err)
	}
}
