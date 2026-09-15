//go:build linux

package trash

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestRenameNoReplaceFallsBackWithoutOverwriting(t *testing.T) {
	root := t.TempDir()
	renameUnsupported := func(string, string) error { return unix.EINVAL }

	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	if err := os.WriteFile(source, []byte("moved"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := renameNoReplace(source, destination, renameUnsupported); err != nil {
		t.Fatalf("renameNoReplace() error = %v", err)
	}
	if content, err := os.ReadFile(destination); err != nil || string(content) != "moved" {
		t.Fatalf("destination content = %q, error = %v", content, err)
	}

	source = filepath.Join(root, "another-source")
	if err := os.WriteFile(source, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := renameNoReplace(source, destination, renameUnsupported); !errors.Is(err, os.ErrExist) {
		t.Fatalf("renameNoReplace() error = %v, want os.ErrExist", err)
	}
	if content, err := os.ReadFile(destination); err != nil || string(content) != "moved" {
		t.Fatalf("existing destination content = %q, error = %v", content, err)
	}
}
