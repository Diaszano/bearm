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

func TestRootResolverUsesHomeTrashWhenPerMountIsDisabled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	resolver := linuxtrash.RootResolver{
		HomeTrash: filepath.Join(root, "home-trash"),
		UID:       1000,
		PerMount:  false,
	}

	got, err := resolver.Resolve(target)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got.Path != resolver.HomeTrash {
		t.Fatalf("Path = %q, want %q", got.Path, resolver.HomeTrash)
	}
}

func TestRootResolverFallsBackToHomeTrashOnUnwritableMount(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("Skipping test because root can write to /")
	}

	root := t.TempDir()
	resolver := linuxtrash.RootResolver{
		HomeTrash: filepath.Join(root, "home-trash"),
		UID:       os.Getuid(),
		PerMount:  true,
	}

	// "/usr/bin" is on a different filesystem (usually /), which is read-only/unwritable for non-root users.
	got, err := resolver.Resolve("/usr/bin")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	// Verify that it resolved either to home trash (because we can't write to /)
	// or, if run as root, it could create the root trash.
	if got.Path != resolver.HomeTrash {
		t.Fatalf("got %q, want %q", got.Path, resolver.HomeTrash)
	}
}

func TestRootEnsureRejectsSymlink(t *testing.T) {
	t.Parallel()

	rootPath := filepath.Join(t.TempDir(), "TrashSymlink")
	targetDir := t.TempDir()
	if err := os.Symlink(targetDir, rootPath); err != nil {
		t.Fatal(err)
	}

	root := linuxtrash.Root{Path: rootPath}
	if err := root.Ensure(); err == nil {
		t.Fatal("Ensure() succeeded, want error for symlink")
	}
}
