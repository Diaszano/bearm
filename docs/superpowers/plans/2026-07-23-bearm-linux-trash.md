# Bearm Linux Trash Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement an atomic, collision-safe Linux trash backend that follows the FreeDesktop Trash specification and uses per-mount trash by default.

**Architecture:** Linux-specific device and mount helpers live behind `//go:build linux`. The backend resolves the correct home or mount trash root, reserves `.trashinfo` metadata with `O_EXCL`, renames the source, and rolls back metadata on failure. Dependencies such as clock and ID generation are injected for deterministic tests.

**Tech Stack:** Go 1.24+, Go standard library, `golang.org/x/sys/unix`, Linux integration tests.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- No shell execution or external GNU utility dependency.
- Follow FreeDesktop trash directory and `.trashinfo` semantics.
- Per-mount trash routing is enabled by default.
- Never overwrite existing trash files or metadata.
- Use atomic exclusive metadata creation.
- A failed move must leave the source untouched and remove newly reserved metadata.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
internal/
├── platform/
│   ├── device_linux.go
│   └── device_linux_test.go
└── trash/
    ├── encoding.go
    ├── encoding_test.go
    ├── linux/
    │   ├── backend.go
    │   ├── backend_test.go
    │   ├── integration_test.go
    │   ├── metadata.go
    │   ├── metadata_test.go
    │   ├── mount.go
    │   ├── mount_test.go
    │   ├── reserve.go
    │   └── reserve_test.go
    └── testdata/
        └── trashinfo.golden
```

## Dependency Order

```text
Task 1 path encoding and metadata
  ├── Task 2 Linux device and mount resolution
  ├── Task 3 collision-safe reservation
  └── Task 4 backend move transaction
          └── Task 5 integration and concurrency tests
```

### Task 1: Encode FreeDesktop Paths and Render Trash Metadata

**Files:**
- Create: `internal/trash/encoding.go`
- Create: `internal/trash/encoding_test.go`
- Create: `internal/trash/linux/metadata.go`
- Create: `internal/trash/linux/metadata_test.go`

**Interfaces:**
- Consumes: UTF-8 filesystem path and deletion time.
- Produces:
  - `trash.EncodePath(string) string`
  - `linux.RenderTrashInfo(path string, deletedAt time.Time) []byte`

- [ ] **Step 1: Write failing percent-encoding tests**

Create `internal/trash/encoding_test.go`:

```go
package trash_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/trash"
)

func TestEncodePathPreservesSafeCharacters(t *testing.T) {
	t.Parallel()

	input := "/home/dias/project/file-name_1.txt"
	if got := trash.EncodePath(input); got != input {
		t.Fatalf("EncodePath() = %q", got)
	}
}

func TestEncodePathEscapesWhitespaceAndPercent(t *testing.T) {
	t.Parallel()

	input := "/home/dias/a b%/ç.txt"
	want := "/home/dias/a%20b%25/%C3%A7.txt"
	if got := trash.EncodePath(input); got != want {
		t.Fatalf("EncodePath() = %q, want %q", got, want)
	}
}

func TestEncodePathUsesUppercaseHex(t *testing.T) {
	t.Parallel()

	if got := trash.EncodePath("\n"); got != "%0A" {
		t.Fatalf("EncodePath() = %q", got)
	}
}
```

- [ ] **Step 2: Write failing metadata tests**

Create `internal/trash/linux/metadata_test.go`:

```go
package linux_test

import (
	"testing"
	"time"

	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestRenderTrashInfo(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, 7, 23, 9, 30, 15, 0, time.FixedZone("BRT", -3*60*60))
	got := string(linuxtrash.RenderTrashInfo("/home/dias/a b.txt", deletedAt))
	want := "[Trash Info]\nPath=/home/dias/a%20b.txt\nDeletionDate=2026-07-23T09:30:15\n"

	if got != want {
		t.Fatalf("RenderTrashInfo() = %q, want %q", got, want)
	}
}
```

- [ ] **Step 3: Run tests and verify failure**

Run on Linux:

```bash
go test ./internal/trash/... -run 'TestEncodePath|TestRenderTrashInfo' -v
```

Expected:

```text
FAIL because EncodePath and RenderTrashInfo do not exist.
```

- [ ] **Step 4: Implement byte-safe path encoding**

Create `internal/trash/encoding.go`:

```go
// Package trash contains platform-independent trash helpers.
package trash

import "strings"

const uppercaseHex = "0123456789ABCDEF"

// EncodePath percent-encodes a UTF-8 path for a FreeDesktop .trashinfo file.
func EncodePath(path string) string {
	var builder strings.Builder
	builder.Grow(len(path))

	for index := 0; index < len(path); index++ {
		value := path[index]
		if isSafePathByte(value) {
			builder.WriteByte(value)
			continue
		}

		builder.WriteByte('%')
		builder.WriteByte(uppercaseHex[value>>4])
		builder.WriteByte(uppercaseHex[value&0x0f])
	}

	return builder.String()
}

func isSafePathByte(value byte) bool {
	switch {
	case value >= 'a' && value <= 'z':
		return true
	case value >= 'A' && value <= 'Z':
		return true
	case value >= '0' && value <= '9':
		return true
	}

	switch value {
	case '-', '_', '.', '~', '/':
		return true
	default:
		return false
	}
}
```

- [ ] **Step 5: Implement `.trashinfo` rendering**

Create `internal/trash/linux/metadata.go`:

```go
//go:build linux

package linux

import (
	"fmt"
	"time"

	"github.com/Diaszano/bearm/internal/trash"
)

// RenderTrashInfo renders a FreeDesktop .trashinfo document.
func RenderTrashInfo(path string, deletedAt time.Time) []byte {
	return []byte(fmt.Sprintf(
		"[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		trash.EncodePath(path),
		deletedAt.Format("2006-01-02T15:04:05"),
	))
}
```

- [ ] **Step 6: Run tests**

Run:

```bash
gofmt -w internal/trash
go test ./internal/trash/... -run 'TestEncodePath|TestRenderTrashInfo' -v
```

Expected:

```text
PASS
```

- [ ] **Step 7: Commit**

```bash
git add internal/trash
git commit -m "feat: render FreeDesktop trash metadata"
```

### Task 2: Resolve Linux Device IDs and Trash Roots

**Files:**
- Modify: `go.mod`
- Modify: `go.sum`
- Create: `internal/platform/device_linux.go`
- Create: `internal/platform/device_linux_test.go`
- Create: `internal/trash/linux/mount.go`
- Create: `internal/trash/linux/mount_test.go`

**Interfaces:**
- Consumes: target path, home trash path, user ID, XDG data path.
- Produces:
  - `platform.DeviceID(path string) (uint64, error)`
  - `platform.MountPoint(path string) (string, error)`
  - `linux.RootResolver`
  - `linux.Root`
  - `(*RootResolver).Resolve(targetPath string) (Root, error)`

- [ ] **Step 1: Add the Linux syscall dependency**

Run:

```bash
go get golang.org/x/sys@v0.35.0
go mod tidy
```

Expected:

```text
go.mod includes golang.org/x/sys and go.sum is updated.
```

- [ ] **Step 2: Write failing platform tests**

Create `internal/platform/device_linux_test.go`:

```go
//go:build linux

package platform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/platform"
)

func TestDeviceIDMatchesParentFilesystem(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	rootID, err := platform.DeviceID(root)
	if err != nil {
		t.Fatal(err)
	}
	fileID, err := platform.DeviceID(file)
	if err != nil {
		t.Fatal(err)
	}
	if rootID != fileID {
		t.Fatalf("root device = %d, file device = %d", rootID, fileID)
	}
}

func TestMountPointReturnsAbsoluteExistingDirectory(t *testing.T) {
	t.Parallel()

	got, err := platform.MountPoint(t.TempDir())
	if err != nil {
		t.Fatalf("MountPoint() error = %v", err)
	}
	info, err := os.Stat(got)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", got, err)
	}
	if !info.IsDir() {
		t.Fatalf("mount point %q is not a directory", got)
	}
}
```

- [ ] **Step 3: Implement Linux device helpers**

Create `internal/platform/device_linux.go`:

```go
//go:build linux

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

- [ ] **Step 4: Write failing trash-root tests**

Create `internal/trash/linux/mount_test.go`:

```go
//go:build linux

package linux_test

import (
	"os"
	"path/filepath"
	"testing"

	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestRootResolverUsesHomeTrashOnSameDevice(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	resolver := linuxtrash.RootResolver{
		HomeTrash: filepath.Join(root, "home-trash"),
		UID:       1000,
		PerMount:  true,
	}

	got, err := resolver.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Path != resolver.HomeTrash {
		t.Fatalf("Path = %q, want %q", got.Path, resolver.HomeTrash)
	}
	if got.RelativeInfoPath {
		t.Fatal("RelativeInfoPath = true, want false")
	}
}

func TestRootEnsureCreatesFilesAndInfo(t *testing.T) {
	t.Parallel()

	root := linuxtrash.Root{Path: filepath.Join(t.TempDir(), "Trash")}
	if err := root.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	for _, name := range []string{"files", "info"} {
		info, err := os.Stat(filepath.Join(root.Path, name))
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", name, err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", name)
		}
	}
}
```

- [ ] **Step 5: Implement root resolution**

Create `internal/trash/linux/mount.go`:

```go
//go:build linux

package linux

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/platform"
)

// Root is one FreeDesktop trash root.
type Root struct {
	Path             string
	MountPoint       string
	RelativeInfoPath bool
}

// Ensure creates and validates the trash skeleton.
func (r Root) Ensure() error {
	if !filepath.IsAbs(r.Path) {
		return errors.New("trash root must be absolute")
	}

	if err := ensureRealDirectory(r.Path, 0o700); err != nil {
		return err
	}
	if err := ensureRealDirectory(filepath.Join(r.Path, "files"), 0o700); err != nil {
		return err
	}
	return ensureRealDirectory(filepath.Join(r.Path, "info"), 0o700)
}

// RootResolver selects the same-filesystem trash root for a target.
type RootResolver struct {
	HomeTrash string
	UID       int
	PerMount  bool
}

// Resolve selects the FreeDesktop trash root for targetPath.
func (r RootResolver) Resolve(targetPath string) (Root, error) {
	if !filepath.IsAbs(r.HomeTrash) {
		return Root{}, errors.New("home trash must be absolute")
	}

	targetDevice, err := platform.DeviceID(targetPath)
	if err != nil {
		return Root{}, err
	}

	homeParent := filepath.Dir(r.HomeTrash)
	if err := os.MkdirAll(homeParent, 0o700); err != nil {
		return Root{}, err
	}
	homeDevice, err := platform.DeviceID(homeParent)
	if err != nil {
		return Root{}, err
	}

	if !r.PerMount || targetDevice == homeDevice {
		root := Root{Path: r.HomeTrash}
		return root, root.Ensure()
	}

	mountPoint, err := platform.MountPoint(targetPath)
	if err != nil {
		return Root{}, err
	}

	adminTrash := filepath.Join(mountPoint, ".Trash")
	if validAdminTrash(adminTrash) {
		root := Root{
			Path:             filepath.Join(adminTrash, fmt.Sprintf("%d", r.UID)),
			MountPoint:       mountPoint,
			RelativeInfoPath: true,
		}
		if err := root.Ensure(); err == nil {
			return root, nil
		}
	}

	root := Root{
		Path:             filepath.Join(mountPoint, fmt.Sprintf(".Trash-%d", r.UID)),
		MountPoint:       mountPoint,
		RelativeInfoPath: true,
	}
	if err := root.Ensure(); err != nil {
		return Root{}, err
	}
	return root, nil
}

func ensureRealDirectory(path string, mode os.FileMode) error {
	info, err := os.Lstat(path)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("%s is not a real directory", path)
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(path, mode)
}

func validAdminTrash(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	return info.Mode()&os.ModeSticky != 0
}
```

- [ ] **Step 6: Run platform and root tests**

Run:

```bash
gofmt -w internal/platform internal/trash/linux
go test ./internal/platform ./internal/trash/linux -run 'TestDeviceID|TestMountPoint|TestRoot' -v
```

Expected:

```text
PASS
```

- [ ] **Step 7: Commit**

```bash
git add go.mod go.sum internal/platform internal/trash/linux/mount.go internal/trash/linux/mount_test.go
git commit -m "feat: resolve Linux trash roots"
```

### Task 3: Reserve Collision-Safe Trash Names

**Files:**
- Create: `internal/trash/linux/reserve.go`
- Create: `internal/trash/linux/reserve_test.go`

**Interfaces:**
- Consumes: `linux.Root`, source base name, metadata bytes.
- Produces:
  - `linux.Reservation`
  - `linux.Reserve(root, base, metadata) (Reservation, error)`
  - `(*Reservation).Rollback() error`

- [ ] **Step 1: Write failing reservation tests**

Create `internal/trash/linux/reserve_test.go`:

```go
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
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/trash/linux -run 'TestReserve|TestReservation' -v
```

Expected:

```text
FAIL because Reserve and Reservation do not exist.
```

- [ ] **Step 3: Implement atomic reservation**

Create `internal/trash/linux/reserve.go`:

```go
//go:build linux

package linux

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Reservation owns one exclusively reserved metadata name.
type Reservation struct {
	TargetPath string
	InfoPath   string
	active     bool
}

// Reserve atomically reserves a matching files/ and info/ name.
func Reserve(root Root, base string, metadata []byte) (Reservation, error) {
	if base == "" || base == "." || base == ".." || filepath.Base(base) != base {
		return Reservation{}, errors.New("invalid trash base name")
	}

	for suffix := 0; suffix < 10000; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", base, suffix)
		}

		targetPath := filepath.Join(root.Path, "files", candidate)
		infoPath := filepath.Join(root.Path, "info", candidate+".trashinfo")

		if pathExists(targetPath) {
			continue
		}

		file, err := os.OpenFile(infoPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return Reservation{}, err
		}

		writeErr := writeAndSync(file, metadata)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(infoPath)
			return Reservation{}, writeErr
		}
		if closeErr != nil {
			_ = os.Remove(infoPath)
			return Reservation{}, closeErr
		}

		return Reservation{
			TargetPath: targetPath,
			InfoPath:   infoPath,
			active:     true,
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
	err := os.Remove(r.InfoPath)
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

- [ ] **Step 4: Run reservation tests**

Run:

```bash
gofmt -w internal/trash/linux
go test ./internal/trash/linux -run 'TestReserve|TestReservation' -v
```

Expected:

```text
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/trash/linux/reserve.go internal/trash/linux/reserve_test.go
git commit -m "feat: reserve Linux trash names atomically"
```

### Task 4: Implement the Linux Backend Move Transaction

**Files:**
- Create: `internal/trash/linux/backend.go`
- Create: `internal/trash/linux/backend_test.go`

**Interfaces:**
- Consumes:
  - `domain.PlannedTarget`
  - `linux.RootResolver`
  - injected clock and ID generator.
- Produces:
  - `linux.Backend`
  - `linux.NewBackend(resolver, clock, idGenerator)`
  - implementation of `domain.TrashBackend`.

- [ ] **Step 1: Write failing backend tests**

Create `internal/trash/linux/backend_test.go`:

```go
//go:build linux

package linux_test

import (
	"context"
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
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/trash/linux -run TestBackendMove -v
```

Expected:

```text
FAIL because NewBackend does not exist.
```

- [ ] **Step 3: Implement the backend**

Create `internal/trash/linux/backend.go`:

```go
//go:build linux

package linux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Backend moves Linux targets into FreeDesktop trash.
type Backend struct {
	resolver    RootResolver
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a Linux trash backend.
func NewBackend(resolver RootResolver, clock func() time.Time, idGenerator func() (string, error)) *Backend {
	return &Backend{
		resolver:    resolver,
		clock:       clock,
		idGenerator: idGenerator,
	}
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "linux-freedesktop"
}

// Resolve selects and prepares the trash root for target.
func (b *Backend) Resolve(_ context.Context, target domain.PlannedTarget) (domain.Destination, error) {
	root, err := b.resolver.Resolve(target.AbsolutePath)
	if err != nil {
		return domain.Destination{}, err
	}

	return domain.Destination{
		Root:     root.Path,
		FilesDir: filepath.Join(root.Path, "files"),
		InfoDir:  filepath.Join(root.Path, "info"),
	}, nil
}

// Move atomically moves target into its resolved trash root.
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

	root := Root{Path: destination.Root}
	if err := root.Ensure(); err != nil {
		return domain.TrashRecord{}, err
	}

	deletedAt := b.clock()
	infoPathValue := target.AbsolutePath
	if resolvedRoot, err := b.resolver.Resolve(target.AbsolutePath); err == nil && resolvedRoot.RelativeInfoPath {
		relative, relErr := filepath.Rel(resolvedRoot.MountPoint, target.AbsolutePath)
		if relErr != nil {
			return domain.TrashRecord{}, relErr
		}
		infoPathValue = relative
	}

	reservation, err := Reserve(
		root,
		filepath.Base(target.AbsolutePath),
		RenderTrashInfo(infoPathValue, deletedAt.Local()),
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

- [ ] **Step 4: Run backend tests**

Run:

```bash
gofmt -w internal/trash/linux
go test ./internal/trash/linux -run TestBackendMove -v
```

Expected:

```text
PASS
```

- [ ] **Step 5: Verify interface compliance**

Append to `internal/trash/linux/backend_test.go`:

```go
var _ domain.TrashBackend = (*linuxtrash.Backend)(nil)
```

Run:

```bash
go test ./internal/trash/linux -v
```

Expected:

```text
PASS
```

- [ ] **Step 6: Commit**

```bash
git add internal/trash/linux/backend.go internal/trash/linux/backend_test.go
git commit -m "feat: move Linux targets to trash"
```

### Task 5: Add Linux Integration and Concurrency Tests

**Files:**
- Create: `internal/trash/linux/integration_test.go`
- Modify: `internal/trash/linux/reserve.go`

**Interfaces:**
- Consumes: completed Linux backend.
- Produces: verified behavior for symlinks, directories, Unicode paths, and concurrent name collisions.

- [ ] **Step 1: Add integration tests**

Create `internal/trash/linux/integration_test.go`:

```go
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
```

- [ ] **Step 2: Run integration and race tests**

Run on Linux:

```bash
go test ./internal/trash/linux -v
go test -race ./internal/trash/linux -run TestConcurrentMovesReceiveUniqueNames -count=10
```

Expected:

```text
PASS with no race reports.
```

- [ ] **Step 3: Harden reservation against a target created after metadata reservation**

Modify `Reserve` in `internal/trash/linux/reserve.go` immediately after the metadata file is closed:

```go
if pathExists(targetPath) {
	_ = os.Remove(infoPath)
	continue
}
```

This closes the ordinary race where another actor creates the target path between the initial existence check and metadata reservation.

- [ ] **Step 4: Re-run concurrency tests**

Run:

```bash
go test -race ./internal/trash/linux -count=10
```

Expected:

```text
PASS with no duplicate destination and no race reports.
```

- [ ] **Step 5: Run Linux package verification**

Run:

```bash
gofmt -w internal/trash/linux
go vet ./internal/trash/...
go test ./internal/trash/...
go test -race ./internal/trash/...
```

Expected:

```text
All commands exit 0.
```

- [ ] **Step 6: Commit**

```bash
git add internal/trash/linux
git commit -m "test: harden Linux trash concurrency"
```

## Plan Completion Verification

Run on Linux:

```bash
make verify
go test -race ./internal/trash/linux -count=10
```

Expected:

```text
All commands succeed.
Linux files, directories, Unicode names, and dangling symlinks move to
collision-safe FreeDesktop trash locations.
```
