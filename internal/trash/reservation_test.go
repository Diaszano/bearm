package trash_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/trash"
)

func TestReserveNameUsesMatchingFilesAndMetadataNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filesDir := filepath.Join(root, "files")
	metadataDir := filepath.Join(root, "info")
	for _, directory := range []string{filesDir, metadataDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	reservation, err := trash.ReserveName(
		filesDir,
		metadataDir,
		"file.txt",
		".trashinfo",
		[]byte("metadata"),
	)
	if err != nil {
		t.Fatalf("ReserveName() error = %v", err)
	}
	t.Cleanup(func() { _ = reservation.Rollback() })

	if filepath.Base(reservation.TargetPath) != "file.txt" {
		t.Fatalf("TargetPath = %q", reservation.TargetPath)
	}
	if filepath.Base(reservation.MetadataPath) != "file.txt.trashinfo" {
		t.Fatalf("MetadataPath = %q", reservation.MetadataPath)
	}
}

func TestReserveNameRetriesMetadataCollision(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filesDir := filepath.Join(root, "files")
	metadataDir := filepath.Join(root, "info")
	for _, directory := range []string{filesDir, metadataDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(filepath.Join(filesDir, "file.txt"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(metadataDir, "file.txt.1.trashinfo"), []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}

	reservation, err := trash.ReserveName(
		filesDir,
		metadataDir,
		"file.txt",
		".trashinfo",
		[]byte("new"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = reservation.Rollback() })

	if filepath.Base(reservation.TargetPath) != "file.txt.2" {
		t.Fatalf("TargetPath = %q", reservation.TargetPath)
	}
}

func TestReservationRollbackAndCommit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filesDir := filepath.Join(root, "files")
	metadataDir := filepath.Join(root, "info")
	for _, directory := range []string{filesDir, metadataDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}

	reservation, err := trash.ReserveName(filesDir, metadataDir, "file.txt", ".trashinfo", []byte("metadata"))
	if err != nil {
		t.Fatal(err)
	}
	if err := reservation.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reservation.MetadataPath); !os.IsNotExist(err) {
		t.Fatalf("Stat(metadata) error = %v, want not exist after Rollback", err)
	}

	reservation2, err := trash.ReserveName(filesDir, metadataDir, "file.txt", ".trashinfo", []byte("metadata"))
	if err != nil {
		t.Fatal(err)
	}
	reservation2.Commit()
	if err := reservation2.Rollback(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(reservation2.MetadataPath); err != nil {
		t.Fatalf("Stat(metadata) error = %v, want file to exist after Commit", err)
	}
}

func TestReserveNameRejectsUnsafeBase(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	invalids := []string{"", ".", "..", "/", "../file", "foo/bar", "a/b/c"}
	for _, invalid := range invalids {
		if _, err := trash.ReserveName(root, root, invalid, ".meta", nil); err == nil {
			t.Errorf("ReserveName() error = nil for base %q, want non-nil", invalid)
		}
	}
}

func TestReserveNameRejectsInvalidSuffix(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	invalids := []string{"", "foo/bar", "../suffix"}
	for _, invalid := range invalids {
		if _, err := trash.ReserveName(root, root, "file.txt", invalid, nil); err == nil {
			t.Errorf("ReserveName() error = nil for suffix %q, want non-nil", invalid)
		}
	}
}
