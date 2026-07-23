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
