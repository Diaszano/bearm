//go:build linux

package linux_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestBackendMoveRenamesFileAndReturnsRecord(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(root, "Trash"),
			UID:       os.Getuid(),
			PerMount:  true,
		},
		func() time.Time { return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC) },
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
		DeviceID:     1,
	}

	destination, err := backend.Resolve(context.Background(), target)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	record, err := backend.Move(context.Background(), target, destination, "operation-1")
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}

	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	if got, err := os.ReadFile(record.TrashedPath); err != nil || string(got) != "data" {
		t.Fatalf("trashed content = %q, error = %v", got, err)
	}
	if record.ItemID != "item-1" || record.OperationID != "operation-1" {
		t.Fatalf("record = %#v", record)
	}
}

func TestBackendMoveRollsBackMetadataWhenRenameFails(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "missing.txt")
	trashRoot := filepath.Join(root, "Trash")

	backend := linuxtrash.NewBackend(
		linuxtrash.RootResolver{HomeTrash: trashRoot, UID: os.Getuid(), PerMount: true},
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}

	destination := domain.Destination{
		Root:     trashRoot,
		FilesDir: filepath.Join(trashRoot, "files"),
		InfoDir:  filepath.Join(trashRoot, "info"),
	}

	if _, err := backend.Move(context.Background(), target, destination, "operation-1"); err == nil {
		t.Fatal("Move() error = nil, want non-nil")
	}

	entries, err := os.ReadDir(filepath.Join(trashRoot, "info"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("metadata entries = %d, want 0", len(entries))
	}
}

var _ domain.TrashBackend = (*linuxtrash.Backend)(nil)

func TestBackendMoveFailsWithEmptyOperationID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := linuxtrash.NewBackend(
		linuxtrash.RootResolver{HomeTrash: filepath.Join(root, "Trash"), UID: os.Getuid(), PerMount: true},
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(root, "Trash"),
		FilesDir: filepath.Join(root, "Trash", "files"),
		InfoDir:  filepath.Join(root, "Trash", "info"),
	}

	if _, err := backend.Move(context.Background(), target, destination, ""); err == nil {
		t.Fatal("expected error with empty operation ID, got nil")
	}
}

func TestBackendMoveFailsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := linuxtrash.NewBackend(
		linuxtrash.RootResolver{HomeTrash: filepath.Join(root, "Trash"), UID: os.Getuid(), PerMount: true},
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(root, "Trash"),
		FilesDir: filepath.Join(root, "Trash", "files"),
		InfoDir:  filepath.Join(root, "Trash", "info"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := backend.Move(ctx, target, destination, "op-1"); err == nil {
		t.Fatal("expected error for canceled context, got nil")
	}
}

func TestBackendMoveRevertsOnIDGeneratorFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	idErr := errors.New("id generation failed")
	backend := linuxtrash.NewBackend(
		linuxtrash.RootResolver{HomeTrash: filepath.Join(root, "Trash"), UID: os.Getuid(), PerMount: true},
		time.Now,
		func() (string, error) { return "", idErr },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(root, "Trash"),
		FilesDir: filepath.Join(root, "Trash", "files"),
		InfoDir:  filepath.Join(root, "Trash", "info"),
	}

	if _, err := backend.Move(context.Background(), target, destination, "op-1"); !errors.Is(err, idErr) {
		t.Fatalf("expected error %v, got %v", idErr, err)
	}

	// Verify the source file remains untouched at its original location!
	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("source file was modified or deleted: %v", err)
	}

	// Verify no metadata was left in the trash directory
	entries, err := os.ReadDir(filepath.Join(root, "Trash", "info"))
	if err == nil && len(entries) > 0 {
		t.Fatalf("metadata was not cleaned up: %d files remain", len(entries))
	}
}
