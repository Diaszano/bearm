//go:build linux

package linux_test

import (
	"os"
	"path/filepath"
	"testing"

	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestReserveUsesBaseNameWhenAvailable(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatal(err)
	}

	reservation, err := linuxtrash.Reserve(root, "file.txt", []byte("metadata"))
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	t.Cleanup(func() { _ = reservation.Rollback() })

	if filepath.Base(reservation.TargetPath) != "file.txt" {
		t.Fatalf("TargetPath = %q", reservation.TargetPath)
	}
	if filepath.Base(reservation.InfoPath) != "file.txt.trashinfo" {
		t.Fatalf("InfoPath = %q", reservation.InfoPath)
	}
}

func TestReserveIncrementsOnFileOrMetadataCollision(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(root.Path, "files", "file.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root.Path, "info", "file.txt.1.trashinfo"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	reservation, err := linuxtrash.Reserve(root, "file.txt", []byte("metadata"))
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}
	t.Cleanup(func() { _ = reservation.Rollback() })

	if filepath.Base(reservation.TargetPath) != "file.txt.2" {
		t.Fatalf("TargetPath = %q", reservation.TargetPath)
	}
}

func TestReservationRollbackRemovesOnlyMetadata(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatal(err)
	}

	reservation, err := linuxtrash.Reserve(root, "file.txt", []byte("metadata"))
	if err != nil {
		t.Fatal(err)
	}
	if err := reservation.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reservation.InfoPath); !os.IsNotExist(err) {
		t.Fatalf("Stat(info) error = %v, want not exist", err)
	}
}

func TestReservationCommit(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatal(err)
	}

	reservation, err := linuxtrash.Reserve(root, "file.txt", []byte("metadata"))
	if err != nil {
		t.Fatal(err)
	}
	reservation.Commit()
	if err := reservation.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reservation.InfoPath); err != nil {
		t.Fatalf("Stat(info) error = %v, want file to exist after Commit", err)
	}
}

func TestReserveRejectsInvalidBaseNames(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatal(err)
	}

	invalids := []string{"", ".", "..", "/", "foo/bar", "a/b/c"}
	for _, invalid := range invalids {
		_, err := linuxtrash.Reserve(root, invalid, []byte("metadata"))
		if err == nil {
			t.Errorf("expected error for invalid base name %q, got nil", invalid)
		}
	}
}
