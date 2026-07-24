//go:build linux || darwin

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestCheckSecureFileAcceptsPrivateRegularFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("language = \"pt-BR\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := config.CheckSecureFile(path, os.Getuid()); err != nil {
		t.Fatalf("CheckSecureFile() error = %v", err)
	}
}

func TestCheckSecureFileRejectsGroupWritableFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(""), 0o620); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o620); err != nil {
		t.Fatal(err)
	}
	if err := config.CheckSecureFile(path, os.Getuid()); err == nil {
		t.Fatal("CheckSecureFile() error = nil, want non-nil")
	}
}

func TestCheckSecureFileRejectsSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "config.toml")
	if err := os.WriteFile(target, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.toml")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if err := config.CheckSecureFile(link, os.Getuid()); err == nil {
		t.Fatal("CheckSecureFile() error = nil, want non-nil for symlink")
	}
}

func TestCheckSecureFileRejectsWrongOwner(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatal(err)
	}
	// Pass UID + 1 to force mismatch
	if err := config.CheckSecureFile(path, os.Getuid()+1); err == nil {
		t.Fatal("CheckSecureFile() error = nil, want non-nil for wrong owner")
	}
}

func TestCheckSecureFileRejectsNonExistent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.toml")
	if err := config.CheckSecureFile(path, os.Getuid()); err == nil {
		t.Fatal("CheckSecureFile() error = nil, want non-nil for missing file")
	}
}
