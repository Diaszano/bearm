package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestLoadUsesDefaultsWhenFileDoesNotExist(t *testing.T) {
	t.Parallel()

	got, err := config.Load(filepath.Join(t.TempDir(), "missing.toml"), func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	want := config.Default()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestLoadMergesTOMLIntoDefaults(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
language = "en"

[trash]
per_mount = false

[restore]
collision_policy = "rename"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := config.Load(path, func(string) string { return "" })
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Language != "en" || got.Trash.PerMount || got.Restore.CollisionPolicy != "rename" {
		t.Fatalf("Load() = %#v", got)
	}
	if !got.Safety.PreserveRoot {
		t.Fatal("default PreserveRoot was not retained")
	}
}

func TestLoadRejectsUnknownTOMLField(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("unknown = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := config.Load(path, func(string) string { return "" }); err == nil {
		t.Fatal("Load() error = nil, want non-nil")
	}
}

func TestIsMissing(t *testing.T) {
	t.Parallel()

	if !config.IsMissing(os.ErrNotExist) {
		t.Fatal("IsMissing(os.ErrNotExist) = false, want true")
	}
	if config.IsMissing(os.ErrPermission) {
		t.Fatal("IsMissing(os.ErrPermission) = true, want false")
	}
}
