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

func TestMountPointHandlesFilePath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	file := filepath.Join(root, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := platform.MountPoint(file)
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
