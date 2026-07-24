package custom_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	customtrash "github.com/Diaszano/bearm/internal/trash/custom"
)

func TestBackendMovesFileIntoCustomTrash(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "work", "file.txt")
	trashRoot := filepath.Join(root, "custom-trash")
	if err := os.MkdirAll(filepath.Dir(source), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend, err := customtrash.NewBackend(
		trashRoot,
		func() time.Time { return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC) },
		func() (string, error) { return "item-1", nil },
	)
	if err != nil {
		t.Fatal(err)
	}

	target := domain.PlannedTarget{
		InputPath: source, AbsolutePath: source, Kind: domain.TargetFile,
	}
	destination, err := backend.Resolve(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	record, err := backend.Move(context.Background(), target, destination, "operation-1")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(record.TrashedPath); err != nil {
		t.Fatalf("trashed path missing: %v", err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
}

func TestNewBackendRejectsRelativeRoot(t *testing.T) {
	t.Parallel()

	if _, err := customtrash.NewBackend("relative", time.Now, func() (string, error) { return "id", nil }); err == nil {
		t.Fatal("NewBackend() error = nil, want non-nil")
	}
}

func TestNewBackendRejectsNilParameters(t *testing.T) {
	t.Parallel()

	if _, err := customtrash.NewBackend("/absolute", nil, func() (string, error) { return "id", nil }); err == nil {
		t.Fatal("NewBackend() error = nil for nil clock")
	}
	if _, err := customtrash.NewBackend("/absolute", time.Now, nil); err == nil {
		t.Fatal("NewBackend() error = nil for nil idGenerator")
	}
}

func TestBackendResolveRejectsUnsafeSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// Pre-create the custom-trash directory
	trashRoot := filepath.Join(root, "custom-trash")
	if err := os.MkdirAll(trashRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	// Create files dir as a symlink to another directory
	targetDir := filepath.Join(root, "target")
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		t.Fatal(err)
	}
	symlinkPath := filepath.Join(trashRoot, "files")
	if err := os.Symlink(targetDir, symlinkPath); err != nil {
		t.Fatal(err)
	}

	backend, err := customtrash.NewBackend(trashRoot, time.Now, func() (string, error) { return "id", nil })
	if err != nil {
		t.Fatal(err)
	}

	if _, err := backend.Resolve(context.Background(), domain.PlannedTarget{}); err == nil {
		t.Fatal("Resolve() error = nil, want error for symlink files dir")
	}
}

func TestBackendMoveRejectsCanceledContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend, err := customtrash.NewBackend(filepath.Join(root, "trash"), time.Now, func() (string, error) { return "id", nil })
	if err != nil {
		t.Fatal(err)
	}

	destination, err := backend.Resolve(context.Background(), domain.PlannedTarget{})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	target := domain.PlannedTarget{InputPath: source, AbsolutePath: source, Kind: domain.TargetFile}
	if _, err := backend.Move(ctx, target, destination, "op-1"); err == nil {
		t.Fatal("Move() error = nil, want error for canceled context")
	}
}
