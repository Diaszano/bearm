package platform_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/platform"
)

func TestResolveDirsUsesXDGOnLinux(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"XDG_CONFIG_HOME": "/tmp/config",
		"XDG_STATE_HOME":  "/tmp/state",
		"XDG_DATA_HOME":   "/tmp/data",
	}
	getenv := func(key string) string { return env[key] }

	got, err := platform.ResolveDirs(getenv, "/home/dias", "linux")
	if err != nil {
		t.Fatalf("ResolveDirs() error = %v", err)
	}

	if got.ConfigRoot != "/tmp/config/bearm" {
		t.Fatalf("ConfigRoot = %q", got.ConfigRoot)
	}
	if got.StateRoot != "/tmp/state/bearm" {
		t.Fatalf("StateRoot = %q", got.StateRoot)
	}
	if got.DataRoot != "/tmp/data/bearm" {
		t.Fatalf("DataRoot = %q", got.DataRoot)
	}
}

func TestResolveDirsUsesMacApplicationSupport(t *testing.T) {
	t.Parallel()

	got, err := platform.ResolveDirs(func(string) string { return "" }, "/Users/dias", "darwin")
	if err != nil {
		t.Fatalf("ResolveDirs() error = %v", err)
	}

	want := "/Users/dias/Library/Application Support/Bearm"
	if got.ConfigRoot != want {
		t.Fatalf("ConfigRoot = %q, want %q", got.ConfigRoot, want)
	}
	if got.StateRoot != filepath.Join(want, "state") {
		t.Fatalf("StateRoot = %q", got.StateRoot)
	}
	if got.DataRoot != filepath.Join(want, "data") {
		t.Fatalf("DataRoot = %q", got.DataRoot)
	}
}

func TestResolveDirsRejectsRelativeOverride(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_STATE_HOME": "relative/path"}
	_, err := platform.ResolveDirs(func(key string) string { return env[key] }, "/home/dias", "linux")
	if err == nil {
		t.Fatal("ResolveDirs() error = nil, want non-nil")
	}
}

func TestResolveDirsRejectsInvalidHome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		home string
	}{
		{"empty home", ""},
		{"relative home", "relative/home"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := platform.ResolveDirs(func(string) string { return "" }, tc.home, "linux")
			if err == nil {
				t.Fatal("ResolveDirs() error = nil, want non-nil")
			}
		})
	}
}
