# Bearm macOS Trash Backend Implementation Plan

**Status:** Complete
**Completion date:** 2026-07-25
**Implementation range:** `ac35fabc68158fb381d56a274b9238e110190531..01931408c25b55e95ae4dd12061ce11388b348f8`

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a macOS backend that moves targets into system-visible home or per-volume Trash directories, records Bearm restore metadata, and never follows symlinks.

**Architecture:** Shared trash-name reservation is extracted from the Linux backend and reused with backend-specific metadata directories. macOS mount selection is injected behind small functions for deterministic tests. Version 1 uses direct `rename` operations and Bearm restore metadata rather than Finder scripting.

**Tech Stack:** Go 1.24+, Go standard library, `golang.org/x/sys/unix`, macOS integration tests.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- No AppleScript, shell execution, or external utility dependency in version 1.
- Use `$HOME/.Trash` for the home volume.
- Use `<mount>/.Trashes/<uid>` for other macOS volumes.
- Finder **Put Back** metadata is not guaranteed; Bearm restore is authoritative.
- Never follow or resolve the final symlink operand.
- Never overwrite existing trash items or Bearm metadata.
- Custom cross-filesystem trash moves fail without copying or deleting the source.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
internal/
├── platform/
│   ├── device_darwin.go
│   └── device_darwin_test.go
└── trash/
    ├── reservation.go
    ├── reservation_test.go
    ├── linux/
    │   ├── backend.go
    │   ├── reserve.go        # removed after migration
    │   └── reserve_test.go   # removed after migration
    └── darwin/
        ├── backend.go
        ├── backend_test.go
        ├── integration_test.go
        ├── metadata.go
        ├── metadata_test.go
        ├── mount.go
        └── mount_test.go
```

## Dependency Order

```text
Task 1 shared reservation extraction
  ├── Task 2 Darwin mount resolution
  ├── Task 3 Darwin metadata and backend
  └── Task 4 integration, race, and CI verification
```

### Task 1: Extract Cross-Platform Trash Name Reservation

**Files:**
- Create: `internal/trash/reservation.go`
- Create: `internal/trash/reservation_test.go`
- Modify: `internal/trash/linux/backend.go`
- Delete: `internal/trash/linux/reserve.go`
- Delete: `internal/trash/linux/reserve_test.go`

**Interfaces:**
- Consumes: files directory, metadata directory, base name, metadata suffix, metadata bytes.
- Produces:
  - `trash.Reservation`
  - `trash.ReserveName(filesDir, metadataDir, base, metadataSuffix, metadata)`
  - `(*trash.Reservation).Commit()`
  - `(*trash.Reservation).Rollback() error`

- [x] **Step 1: Write shared reservation tests before moving implementation**

Create `internal/trash/reservation_test.go`:

```go
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

	if err := os.WriteFile(filepath.Join(metadataDir, "file.txt.trashinfo"), []byte("old"), 0o600); err != nil {
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

	if filepath.Base(reservation.TargetPath) != "file.txt.1" {
		t.Fatalf("TargetPath = %q", reservation.TargetPath)
	}
}

func TestReserveNameRejectsUnsafeBase(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if _, err := trash.ReserveName(root, root, "../file", ".meta", nil); err == nil {
		t.Fatal("ReserveName() error = nil, want non-nil")
	}
}
```

- [x] **Step 2: Run shared tests and verify failure**

Run:

```bash
go test ./internal/trash -run TestReserveName -v
```

Expected:

```text
FAIL because ReserveName does not exist.
```

- [x] **Step 3: Implement shared reservation**

Create `internal/trash/reservation.go`:

```go
package trash

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Reservation owns one exclusively reserved metadata name.
type Reservation struct {
	TargetPath   string
	MetadataPath string
	active       bool
}

// ReserveName reserves a destination by exclusively creating matching metadata.
func ReserveName(
	filesDir string,
	metadataDir string,
	base string,
	metadataSuffix string,
	metadata []byte,
) (Reservation, error) {
	if base == "" || base == "." || base == ".." || filepath.Base(base) != base {
		return Reservation{}, errors.New("invalid trash base name")
	}
	if metadataSuffix == "" || filepath.Base(metadataSuffix) != metadataSuffix {
		return Reservation{}, errors.New("invalid metadata suffix")
	}

	for suffix := 0; suffix < 10000; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", base, suffix)
		}

		targetPath := filepath.Join(filesDir, candidate)
		metadataPath := filepath.Join(metadataDir, candidate+metadataSuffix)
		if pathExists(targetPath) {
			continue
		}

		file, err := os.OpenFile(metadataPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return Reservation{}, err
		}

		writeErr := writeAndSync(file, metadata)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(metadataPath)
			return Reservation{}, writeErr
		}
		if closeErr != nil {
			_ = os.Remove(metadataPath)
			return Reservation{}, closeErr
		}
		if pathExists(targetPath) {
			_ = os.Remove(metadataPath)
			continue
		}

		return Reservation{
			TargetPath:   targetPath,
			MetadataPath: metadataPath,
			active:       true,
		}, nil
	}

	return Reservation{}, errors.New("trash name reservation limit exceeded")
}

// Commit marks the reservation as completed.
func (r *Reservation) Commit() {
	r.active = false
}

// Rollback removes metadata owned by an incomplete reservation.
func (r *Reservation) Rollback() error {
	if !r.active {
		return nil
	}
	r.active = false
	err := os.Remove(r.MetadataPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil || !os.IsNotExist(err)
}

func writeAndSync(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}
```

- [x] **Step 4: Migrate the Linux backend**

In `internal/trash/linux/backend.go`, add the shared package import if missing:

```go
"github.com/Diaszano/bearm/internal/trash"
```

Replace:

```go
reservation, err := Reserve(
	root,
	filepath.Base(target.AbsolutePath),
	RenderTrashInfo(infoPathValue, deletedAt.Local()),
)
```

with:

```go
reservation, err := trash.ReserveName(
	filepath.Join(root.Path, "files"),
	filepath.Join(root.Path, "info"),
	filepath.Base(target.AbsolutePath),
	".trashinfo",
	RenderTrashInfo(infoPathValue, deletedAt.Local()),
)
```

Replace any use of `reservation.InfoPath` with `reservation.MetadataPath`.

Delete:

```text
internal/trash/linux/reserve.go
internal/trash/linux/reserve_test.go
```

- [x] **Step 5: Run shared and Linux tests**

Run:

```bash
gofmt -w internal/trash
go test ./internal/trash/...
go test -race ./internal/trash/linux -count=10
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add internal/trash
git commit -m "refactor: share atomic trash reservations"
```

### Task 2: Resolve macOS Device IDs and System Trash Roots

**Files:**
- Create: `internal/platform/device_darwin.go`
- Create: `internal/platform/device_darwin_test.go`
- Create: `internal/trash/darwin/mount.go`
- Create: `internal/trash/darwin/mount_test.go`

**Interfaces:**
- Consumes: target path, home directory, UID, injected device and mount functions.
- Produces:
  - `platform.DeviceID(path string) (uint64, error)` on Darwin.
  - `platform.MountPoint(path string) (string, error)` on Darwin.
  - `darwin.RootResolver`
  - `darwin.Root`
  - `(*RootResolver).Resolve(targetPath string) (Root, error)`

- [x] **Step 1: Write Darwin platform tests**

Create `internal/platform/device_darwin_test.go`:

```go
//go:build darwin

package platform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/platform"
)

func TestDarwinDeviceIDDoesNotFollowFinalSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target")
	link := filepath.Join(root, "link")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	if _, err := platform.DeviceID(link); err != nil {
		t.Fatalf("DeviceID() error = %v", err)
	}
}

func TestDarwinMountPointReturnsAbsoluteDirectory(t *testing.T) {
	t.Parallel()

	got, err := platform.MountPoint(t.TempDir())
	if err != nil {
		t.Fatalf("MountPoint() error = %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("MountPoint() = %q, want absolute", got)
	}
}
```

- [x] **Step 2: Implement Darwin device helpers**

Create `internal/platform/device_darwin.go`:

```go
//go:build darwin

package platform

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// DeviceID returns the filesystem device ID for path without following a final symlink.
func DeviceID(path string) (uint64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, errors.New("filesystem stat does not expose a device ID")
	}

	return uint64(stat.Dev), nil
}

// MountPoint returns the top directory on the same device as path.
func MountPoint(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Lstat(absolute)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		absolute = filepath.Dir(absolute)
	}

	device, err := DeviceID(absolute)
	if err != nil {
		return "", err
	}

	current := filepath.Clean(absolute)
	for {
		parent := filepath.Dir(current)
		if parent == current {
			return current, nil
		}
		parentDevice, err := DeviceID(parent)
		if err != nil {
			return "", err
		}
		if parentDevice != device {
			return current, nil
		}
		current = parent
	}
}
```

- [x] **Step 3: Write deterministic root-resolver tests**

Create `internal/trash/darwin/mount_test.go`:

```go
//go:build darwin

package darwin_test

import (
	"os"
	"path/filepath"
	"testing"

	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func TestRootResolverUsesHomeTrashOnHomeDevice(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	target := filepath.Join(home, "project", "file.txt")
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	resolver := darwintrash.RootResolver{
		Home: home,
		UID:  501,
		DeviceID: func(string) (uint64, error) {
			return 10, nil
		},
		MountPoint: func(string) (string, error) {
			return "/", nil
		},
	}

	got, err := resolver.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := filepath.Join(home, ".Trash")
	if got.Path != want {
		t.Fatalf("Path = %q, want %q", got.Path, want)
	}
}

func TestRootResolverUsesPerVolumeTrash(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	external := t.TempDir()
	target := filepath.Join(external, "file.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	resolver := darwintrash.RootResolver{
		Home: home,
		UID:  501,
		DeviceID: func(path string) (uint64, error) {
			if path == home {
				return 10, nil
			}
			return 20, nil
		},
		MountPoint: func(string) (string, error) {
			return external, nil
		},
	}

	got, err := resolver.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	want := filepath.Join(external, ".Trashes", "501")
	if got.Path != want {
		t.Fatalf("Path = %q, want %q", got.Path, want)
	}
}
```

- [x] **Step 4: Implement the macOS root resolver**

Create `internal/trash/darwin/mount.go`:

```go
//go:build darwin

package darwin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/platform"
)

// Root is one macOS trash root.
type Root struct {
	Path string
}

// Ensure creates Bearm metadata storage inside the system-visible trash.
func (r Root) Ensure() error {
	if !filepath.IsAbs(r.Path) {
		return errors.New("trash root must be absolute")
	}
	for _, directory := range []string{
		r.Path,
		filepath.Join(r.Path, ".bearm-info"),
	} {
		info, err := os.Lstat(directory)
		if err == nil {
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("%s is not a real directory", directory)
			}
			continue
		}
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// RootResolver selects a macOS home or per-volume trash.
type RootResolver struct {
	Home       string
	UID        int
	DeviceID   func(string) (uint64, error)
	MountPoint func(string) (string, error)
}

// NewRootResolver creates a resolver using real platform functions.
func NewRootResolver(home string, uid int) RootResolver {
	return RootResolver{
		Home:       home,
		UID:        uid,
		DeviceID:   platform.DeviceID,
		MountPoint: platform.MountPoint,
	}
}

// Resolve selects a same-filesystem system trash for targetPath.
func (r RootResolver) Resolve(targetPath string) (Root, error) {
	if !filepath.IsAbs(r.Home) {
		return Root{}, errors.New("home directory must be absolute")
	}

	targetDevice, err := r.DeviceID(targetPath)
	if err != nil {
		return Root{}, err
	}
	homeDevice, err := r.DeviceID(r.Home)
	if err != nil {
		return Root{}, err
	}

	if targetDevice == homeDevice {
		root := Root{Path: filepath.Join(r.Home, ".Trash")}
		return root, root.Ensure()
	}

	mountPoint, err := r.MountPoint(targetPath)
	if err != nil {
		return Root{}, err
	}
	root := Root{Path: filepath.Join(mountPoint, ".Trashes", fmt.Sprintf("%d", r.UID))}
	return root, root.Ensure()
}
```

- [x] **Step 5: Run tests on macOS**

Run:

```bash
gofmt -w internal/platform internal/trash/darwin
go test ./internal/platform ./internal/trash/darwin -run 'TestDarwin|TestRootResolver' -v
```

Expected:

```text
PASS
```

- [x] **Step 6: Commit**

```bash
git add internal/platform/device_darwin.go internal/platform/device_darwin_test.go internal/trash/darwin/mount.go internal/trash/darwin/mount_test.go
git commit -m "feat: resolve macOS trash roots"
```

### Task 3: Implement macOS Bearm Metadata and Move Backend

**Files:**
- Create: `internal/trash/darwin/metadata.go`
- Create: `internal/trash/darwin/metadata_test.go`
- Create: `internal/trash/darwin/backend.go`
- Create: `internal/trash/darwin/backend_test.go`

**Interfaces:**
- Consumes:
  - `domain.PlannedTarget`
  - `darwin.RootResolver`
  - injected clock and ID generator.
- Produces:
  - `darwin.Metadata`
  - `darwin.RenderMetadata`
  - `darwin.Backend`
  - implementation of `domain.TrashBackend`.

- [x] **Step 1: Write failing metadata tests**

Create `internal/trash/darwin/metadata_test.go`:

```go
//go:build darwin

package darwin_test

import (
	"strings"
	"testing"
	"time"

	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func TestRenderMetadataIsStableJSON(t *testing.T) {
	t.Parallel()

	metadata := darwintrash.Metadata{
		SchemaVersion: 1,
		OriginalPath:  "/Users/dias/a b.txt",
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}

	got, err := darwintrash.RenderMetadata(metadata)
	if err != nil {
		t.Fatalf("RenderMetadata() error = %v", err)
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Fatalf("metadata must end with newline: %q", got)
	}
	if !strings.Contains(string(got), `"original_path":"/Users/dias/a b.txt"`) {
		t.Fatalf("metadata = %q", got)
	}
}
```

- [x] **Step 2: Implement metadata rendering**

Create `internal/trash/darwin/metadata.go`:

```go
//go:build darwin

package darwin

import (
	"bytes"
	"encoding/json"
	"time"
)

// Metadata stores Bearm restore information for a macOS trash item.
type Metadata struct {
	SchemaVersion int       `json:"schema_version"`
	OriginalPath  string    `json:"original_path"`
	DeletedAt     time.Time `json:"deleted_at"`
}

// RenderMetadata renders one newline-terminated JSON metadata document.
func RenderMetadata(metadata Metadata) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(metadata); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
```

- [x] **Step 3: Write failing backend tests**

Create `internal/trash/darwin/backend_test.go`:

```go
//go:build darwin

package darwin_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

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
}
```

- [x] **Step 4: Implement the macOS backend**

Create `internal/trash/darwin/backend.go`:

```go
//go:build darwin

package darwin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

// Backend moves macOS targets into system-visible Trash directories.
type Backend struct {
	resolver    RootResolver
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a macOS trash backend.
func NewBackend(resolver RootResolver, clock func() time.Time, idGenerator func() (string, error)) *Backend {
	return &Backend{resolver: resolver, clock: clock, idGenerator: idGenerator}
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "darwin-system-trash"
}

// Resolve selects and prepares the trash root for target.
func (b *Backend) Resolve(_ context.Context, target domain.PlannedTarget) (domain.Destination, error) {
	root, err := b.resolver.Resolve(target.AbsolutePath)
	if err != nil {
		return domain.Destination{}, err
	}

	return domain.Destination{
		Root:     root.Path,
		FilesDir: root.Path,
		InfoDir:  filepath.Join(root.Path, ".bearm-info"),
	}, nil
}

// Move atomically moves target into its resolved macOS Trash directory.
func (b *Backend) Move(
	ctx context.Context,
	target domain.PlannedTarget,
	destination domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrashRecord{}, err
	}
	if operationID == "" {
		return domain.TrashRecord{}, errors.New("operation ID is required")
	}

	deletedAt := b.clock()
	metadata, err := RenderMetadata(Metadata{
		SchemaVersion: 1,
		OriginalPath:  target.AbsolutePath,
		DeletedAt:     deletedAt.UTC(),
	})
	if err != nil {
		return domain.TrashRecord{}, err
	}

	reservation, err := trash.ReserveName(
		destination.FilesDir,
		destination.InfoDir,
		filepath.Base(target.AbsolutePath),
		".json",
		metadata,
	)
	if err != nil {
		return domain.TrashRecord{}, err
	}
	defer func() { _ = reservation.Rollback() }()

	if err := os.Rename(target.AbsolutePath, reservation.TargetPath); err != nil {
		return domain.TrashRecord{}, err
	}

	itemID, err := b.idGenerator()
	if err != nil {
		return domain.TrashRecord{}, err
	}

	reservation.Commit()
	return domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        itemID,
		OperationID:   operationID,
		OriginalPath:  target.AbsolutePath,
		TrashedPath:   reservation.TargetPath,
		Backend:       b.Name(),
		DeviceID:      target.DeviceID,
		DeletedAt:     deletedAt.UTC(),
		Status:        "trashed",
	}, nil
}
```

- [x] **Step 5: Verify interface compliance**

Append to `internal/trash/darwin/backend_test.go`:

```go
var _ domain.TrashBackend = (*darwintrash.Backend)(nil)
```

- [x] **Step 6: Run macOS backend tests**

Run on macOS:

```bash
gofmt -w internal/trash/darwin
go test ./internal/trash/darwin -v
```

Expected:

```text
PASS
```

- [x] **Step 7: Commit**

```bash
git add internal/trash/darwin
git commit -m "feat: move macOS targets to system trash"
```

### Task 4: Add macOS Integration, Symlink, and Concurrency Tests

**Files:**
- Create: `internal/trash/darwin/integration_test.go`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**
- Consumes: completed macOS backend and shared reservation.
- Produces: verified directory, Unicode, dangling symlink, and collision behavior on macOS CI.

- [x] **Step 1: Add macOS integration tests**

Create `internal/trash/darwin/integration_test.go`:

```go
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
	return record
}
```

- [x] **Step 2: Run macOS integration and race tests**

Run on macOS:

```bash
go test ./internal/trash/darwin -v
go test -race ./internal/trash/darwin -run TestConcurrentDarwinMovesUseUniqueNames -count=10
```

Expected:

```text
PASS with no race reports.
```

- [x] **Step 3: Add backend-specific CI jobs**

Append to `.github/workflows/ci.yml`:

```yaml
  linux-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: go test -race ./internal/trash/linux -count=5

  darwin-backend:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: stable
          cache: true
      - run: go test -race ./internal/trash/darwin -count=5
```

- [x] **Step 4: Run complete platform verification**

Run on macOS:

```bash
make verify
go test -race ./internal/trash/darwin -count=10
```

Run on Linux CI:

```bash
go test -race ./internal/trash/linux -count=10
```

Expected:

```text
All commands succeed.
```

- [x] **Step 5: Commit**

```bash
git add internal/trash/darwin .github/workflows/ci.yml
git commit -m "test: verify macOS trash behavior"
```

## Plan Completion Verification

Run on macOS:

```bash
make verify
go test -race ./internal/trash/darwin -count=10
```

Expected:

```text
Files, directories, Unicode names, and dangling symlinks move to a
system-visible macOS Trash directory with unique Bearm metadata.
```
