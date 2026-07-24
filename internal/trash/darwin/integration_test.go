//go:build darwin

package darwin_test

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
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func TestBackendMovesUnicodeDirectory(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "diretório com espaço")
	if err := os.MkdirAll(filepath.Join(source, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "sub", "ação.txt"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	backend := newDarwinTestBackend(home)
	record := moveDarwinTarget(t, backend, source, domain.TargetDir, "operation-1")

	if _, err := os.Stat(filepath.Join(record.TrashedPath, "sub", "ação.txt")); err != nil {
		t.Fatalf("trashed content missing: %v", err)
	}
}

func TestBackendMovesDanglingSymlinkItself(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	source := filepath.Join(home, "dangling")
	if err := os.Symlink(filepath.Join(home, "missing"), source); err != nil {
		t.Fatal(err)
	}

	backend := newDarwinTestBackend(home)
	record := moveDarwinTarget(t, backend, source, domain.TargetSymlink, "operation-1")

	info, err := os.Lstat(record.TrashedPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("mode = %v, want symlink", info.Mode())
	}
}

func TestConcurrentDarwinMovesUseUniqueNames(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	backend := newDarwinTestBackend(home)

	const count = 20
	var wait sync.WaitGroup
	paths := make(chan string, count)
	errs := make(chan error, count)

	for index := 0; index < count; index++ {
		sourceDir := filepath.Join(home, fmt.Sprintf("source-%d", index))
		if err := os.Mkdir(sourceDir, 0o700); err != nil {
			t.Fatal(err)
		}
		source := filepath.Join(sourceDir, "same.txt")
		if err := os.WriteFile(source, []byte(fmt.Sprintf("%d", index)), 0o600); err != nil {
			t.Fatal(err)
		}

		wait.Add(1)
		go func(path string, operationID string) {
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
			record, err := backend.Move(context.Background(), target, destination, operationID)
			if err != nil {
				errs <- err
				return
			}
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				errs <- fmt.Errorf("source still exists: %v", err)
				return
			}
			paths <- record.TrashedPath
		}(source, fmt.Sprintf("operation-%d", index))
	}

	wait.Wait()
	close(paths)
	close(errs)

	for err := range errs {
		t.Errorf("move error = %v", err)
	}

	seen := map[string]bool{}
	for path := range paths {
		if seen[path] {
			t.Fatalf("duplicate trash path %q", path)
		}
		seen[path] = true
	}
	if len(seen) != count {
		t.Fatalf("unique paths = %d, want %d", len(seen), count)
	}
}

func newDarwinTestBackend(home string) *darwintrash.Backend {
	var sequence atomic.Int64
	return darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		func() time.Time { return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC) },
		func() (string, error) { return fmt.Sprintf("item-%d", sequence.Add(1)), nil },
	)
}

func moveDarwinTarget(
	t *testing.T,
	backend *darwintrash.Backend,
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
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		t.Fatalf("source still exists: %v", err)
	}
	return record
}
