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

func TestRootEnsureCreatesDirectoryAndBearmInfo(t *testing.T) {
	t.Parallel()

	rootPath := filepath.Join(t.TempDir(), "trashroot")
	root := darwintrash.Root{Path: rootPath}
	if err := root.Ensure(); err != nil {
		t.Fatalf("Ensure() error = %v", err)
	}

	infoPath := filepath.Join(rootPath, ".bearm-info")
	info, err := os.Stat(infoPath)
	if err != nil {
		t.Fatalf("Stat(.bearm-info) error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf(".bearm-info is not a directory")
	}
}

func TestRootEnsureRequiresAbsolutePath(t *testing.T) {
	t.Parallel()

	root := darwintrash.Root{Path: "relative/path"}
	if err := root.Ensure(); err == nil {
		t.Fatal("Ensure() expected error for relative path, got nil")
	}
}

func TestRootEnsureFailsIfFileInWay(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	filePath := filepath.Join(temp, "notadir")
	if err := os.WriteFile(filePath, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}

	root := darwintrash.Root{Path: filePath}
	if err := root.Ensure(); err == nil {
		t.Fatal("Ensure() expected error when path is a file, got nil")
	}
}

func TestRootResolverRequiresAbsoluteHome(t *testing.T) {
	t.Parallel()

	resolver := darwintrash.RootResolver{
		Home: "relative/home",
		UID:  501,
	}

	if _, err := resolver.Resolve("/some/path"); err == nil {
		t.Fatal("Resolve() expected error for relative home, got nil")
	}
}
