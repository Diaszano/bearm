//go:build linux

package linux_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/platform"
	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestBackendMovesDanglingSymlinkWithoutTouchingTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "missing-target"), link); err != nil {
		t.Fatal(err)
	}

	backend := newTestBackend(root)
	record := moveTestTarget(t, backend, link, domain.TargetSymlink, "operation-1")

	info, err := os.Lstat(record.TrashedPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("trashed mode = %v, want symlink", info.Mode())
	}
}

func TestBackendMovesDirectoryAsSingleRename(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(filepath.Join(directory, "pkg"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "pkg", "index.js"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := newTestBackend(root)
	record := moveTestTarget(t, backend, directory, domain.TargetDir, "operation-1")

	if _, err := os.Stat(filepath.Join(record.TrashedPath, "pkg", "index.js")); err != nil {
		t.Fatalf("trashed directory content missing: %v", err)
	}
}

func TestConcurrentMovesReceiveUniqueNames(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	backend := newTestBackend(root)

	const count = 20
	var sequence atomic.Int64
	var wait sync.WaitGroup
	records := make(chan domain.TrashRecord, count)
	errs := make(chan error, count)

	for index := 0; index < count; index++ {
		sourceDir := filepath.Join(root, fmt.Sprintf("source-%d", index))
		if err := os.Mkdir(sourceDir, 0o700); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(sourceDir, "same.txt")
		if err := os.WriteFile(source, []byte(fmt.Sprintf("%d", index)), 0o600); err != nil {
			t.Fatal(err)
		}

		wait.Add(1)
		go func(path string) {
			defer wait.Done()

			deviceID, err := platform.DeviceID(path)
			if err != nil {
				errs <- err
				return
			}
			target := domain.PlannedTarget{
				InputPath:    path,
				AbsolutePath: path,
				Kind:         domain.TargetFile,
				DeviceID:     deviceID,
			}
			destination, err := backend.Resolve(context.Background(), target)
			if err != nil {
				errs <- err
				return
			}
			record, err := backend.Move(
				context.Background(),
				target,
				destination,
				fmt.Sprintf("operation-%d", sequence.Add(1)),
			)
			if err != nil {
				errs <- err
				return
			}
			records <- record
		}(source)
	}

	wait.Wait()
	close(records)
	close(errs)

	for err := range errs {
		t.Errorf("concurrent move error = %v", err)
	}

	seen := map[string]bool{}
	for record := range records {
		if seen[record.TrashedPath] {
			t.Fatalf("duplicate path %q", record.TrashedPath)
		}
		seen[record.TrashedPath] = true
	}
	if len(seen) != count {
		t.Fatalf("moved records = %d, want %d", len(seen), count)
	}
}

func newTestBackend(root string) *linuxtrash.Backend {
	var sequence atomic.Int64
	return linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(root, "Trash"),
			UID:       os.Getuid(),
			PerMount:  true,
		},
		func() time.Time { return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC) },
		func() (string, error) { return fmt.Sprintf("item-%d", sequence.Add(1)), nil },
	)
}

func moveTestTarget(
	t *testing.T,
	backend *linuxtrash.Backend,
	path string,
	kind domain.TargetKind,
	operationID string,
) domain.TrashRecord {
	t.Helper()

	deviceID, err := platform.DeviceID(path)
	if err != nil {
		t.Fatal(err)
	}
	target := domain.PlannedTarget{
		InputPath:    path,
		AbsolutePath: path,
		Kind:         kind,
		DeviceID:     deviceID,
	}
	destination, err := backend.Resolve(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	record, err := backend.Move(context.Background(), target, destination, operationID)
	if err != nil {
		t.Fatal(err)
	}
	return record
}

func TestBackendMovesUnicodeFilename(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "données_🚀.txt")
	if err := os.WriteFile(source, []byte("unicode-data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := newTestBackend(root)
	record := moveTestTarget(t, backend, source, domain.TargetFile, "operation-1")

	if _, err := os.Lstat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists")
	}

	if got, err := os.ReadFile(record.TrashedPath); err != nil || string(got) != "unicode-data" {
		t.Fatalf("trashed content = %q, error = %v", got, err)
	}
}
