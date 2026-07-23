//go:build darwin

package darwin_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

var _ domain.TrashBackend = (*darwintrash.Backend)(nil)

func TestBackendMovesFileIntoHomeTrash(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
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
		t.Fatal(err)
	}
	record, err := backend.Move(context.Background(), target, destination, "operation-1")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	if got, err := os.ReadFile(record.TrashedPath); err != nil || string(got) != "data" {
		t.Fatalf("trashed data = %q, error = %v", got, err)
	}
	if record.Backend != "darwin-system-trash" {
		t.Fatalf("Backend = %q", record.Backend)
	}
	if record.ItemID != "item-1" || record.OperationID != "operation-1" {
		t.Fatalf("record = %#v", record)
	}
}

func TestBackendResolveFailsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "file.txt")
	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := backend.Resolve(ctx, target); err == nil {
		t.Fatal("expected error for canceled context in Resolve(), got nil")
	}
}

func TestBackendMoveFailsWhenContextCancelled(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(home, ".Trash"),
		FilesDir: filepath.Join(home, ".Trash"),
		InfoDir:  filepath.Join(home, ".Trash", ".bearm-info"),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := backend.Move(ctx, target, destination, "op-1"); err == nil {
		t.Fatal("expected error for canceled context in Move(), got nil")
	}
}

func TestBackendMoveFailsWithEmptyOperationID(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(home, ".Trash"),
		FilesDir: filepath.Join(home, ".Trash"),
		InfoDir:  filepath.Join(home, ".Trash", ".bearm-info"),
	}

	if _, err := backend.Move(context.Background(), target, destination, ""); err == nil {
		t.Fatal("expected error with empty operation ID, got nil")
	}
}

func TestBackendMoveRevertsOnIDGeneratorFailure(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	idErr := errors.New("id generation failed")
	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		func() (string, error) { return "", idErr },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}
	destination := domain.Destination{
		Root:     filepath.Join(home, ".Trash"),
		FilesDir: filepath.Join(home, ".Trash"),
		InfoDir:  filepath.Join(home, ".Trash", ".bearm-info"),
	}

	if _, err := backend.Move(context.Background(), target, destination, "op-1"); !errors.Is(err, idErr) {
		t.Fatalf("expected error %v, got %v", idErr, err)
	}

	if _, err := os.Lstat(source); err != nil {
		t.Fatalf("source file was modified or deleted: %v", err)
	}

	entries, err := os.ReadDir(destination.InfoDir)
	if err == nil && len(entries) > 0 {
		t.Fatalf("metadata was not cleaned up: %d files remain", len(entries))
	}
}

func TestBackendMoveRollsBackMetadataWhenRenameFails(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "missing.txt")

	backend := darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		func() (string, error) { return "item-1", nil },
	)

	target := domain.PlannedTarget{
		InputPath:    source,
		AbsolutePath: source,
		Kind:         domain.TargetFile,
	}

	destination := domain.Destination{
		Root:     filepath.Join(home, ".Trash"),
		FilesDir: filepath.Join(home, ".Trash"),
		InfoDir:  filepath.Join(home, ".Trash", ".bearm-info"),
	}
	if err := os.MkdirAll(destination.InfoDir, 0o700); err != nil {
		t.Fatal(err)
	}

	if _, err := backend.Move(context.Background(), target, destination, "operation-1"); err == nil {
		t.Fatal("Move() error = nil, want non-nil")
	}

	entries, err := os.ReadDir(destination.InfoDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("metadata entries = %d, want 0", len(entries))
	}
}
