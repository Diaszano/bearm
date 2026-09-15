//go:build linux

package linux

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidAdminTrash(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()

	// 1. Valid admin trash (directory with sticky bit)
	okDir := filepath.Join(tmp, "ok")
	if err := os.Mkdir(okDir, 0777|os.ModeSticky); err != nil {
		t.Fatal(err)
	}
	if !validAdminTrash(okDir) {
		t.Error("expected validAdminTrash to return true for directory with sticky bit")
	}

	// 2. Invalid (directory without sticky bit)
	badDir := filepath.Join(tmp, "bad")
	if err := os.Mkdir(badDir, 0777); err != nil {
		t.Fatal(err)
	}
	if validAdminTrash(badDir) {
		t.Error("expected validAdminTrash to return false for directory without sticky bit")
	}

	// 3. Invalid (regular file)
	file := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if validAdminTrash(file) {
		t.Error("expected validAdminTrash to return false for regular file")
	}

	// 4. Invalid (symlink)
	link := filepath.Join(tmp, "link")
	if err := os.Symlink(okDir, link); err != nil {
		t.Fatal(err)
	}
	if validAdminTrash(link) {
		t.Error("expected validAdminTrash to return false for symlink")
	}
}

func TestEnsureRealDirectoryValidation(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	uid := os.Getuid()

	// 1. Pre-existing directory with correct owner and permissions
	okDir := filepath.Join(tmp, "ok")
	if err := os.Mkdir(okDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := ensureRealDirectory(okDir, uid, true); err != nil {
		t.Errorf("expected success, got err: %v", err)
	}

	// 2. Pre-existing directory with incorrect permissions (e.g. 0755)
	badPermDir := filepath.Join(tmp, "bad_perm")
	if err := os.Mkdir(badPermDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := ensureRealDirectory(badPermDir, uid, true); err == nil {
		t.Error("expected error for directory with 0755 permissions, got nil")
	}

	// 3. Pre-existing directory with incorrect owner
	if err := ensureRealDirectory(okDir, uid+1, true); err == nil {
		t.Error("expected error for directory owned by different UID, got nil")
	}

	// 5. When path is a regular file, verify that it returns an error
	regularFile := filepath.Join(tmp, "regular_file")
	if err := os.WriteFile(regularFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ensureRealDirectory(regularFile, uid, false); err == nil {
		t.Error("expected error when path is a regular file, got nil")
	}
}

func TestRootResolverEdgeCases(t *testing.T) {
	t.Parallel()

	resolver := RootResolver{
		HomeTrash: "relative/path/Trash",
		UID:       os.Getuid(),
		PerMount:  true,
	}
	if _, err := resolver.Resolve("/tmp"); err == nil {
		t.Error("expected error when HomeTrash is not absolute, got nil")
	}

	absResolver := RootResolver{
		HomeTrash: filepath.Join(t.TempDir(), "Trash"),
		UID:       os.Getuid(),
		PerMount:  true,
	}
	if _, err := absResolver.Resolve("/nonexistent/file/path/here"); err == nil {
		t.Error("expected error when targetPath does not exist, got nil")
	}
}

func TestRootEnsureSubpathAsFile(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	rootPath := filepath.Join(tmp, "Trash")
	if err := os.MkdirAll(rootPath, 0700); err != nil {
		t.Fatal(err)
	}
	// Create files path as a regular file
	filesFile := filepath.Join(rootPath, "files")
	if err := os.WriteFile(filesFile, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	root := Root{Path: rootPath, UID: os.Getuid()}
	if err := root.Ensure(); err == nil {
		t.Error("expected error when files is a regular file, got nil")
	}

	// Create info path as a regular file
	tmp2 := t.TempDir()
	rootPath2 := filepath.Join(tmp2, "Trash")
	if err := os.MkdirAll(rootPath2, 0700); err != nil {
		t.Fatal(err)
	}
	infoFile := filepath.Join(rootPath2, "info")
	if err := os.WriteFile(infoFile, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}

	root2 := Root{Path: rootPath2, UID: os.Getuid()}
	if err := root2.Ensure(); err == nil {
		t.Error("expected error when info is a regular file, got nil")
	}
}

func TestRootResolverPerMountShm(t *testing.T) {
	t.Parallel()

	// Check if /dev/shm exists and is writable
	shmDir := "/dev/shm"
	testFile := filepath.Join(shmDir, "bearm_mount_test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Skip("/dev/shm not writable or available")
	}
	defer func() { _ = os.Remove(testFile) }()

	homeTrash := filepath.Join(t.TempDir(), "home-trash")
	resolver := RootResolver{
		HomeTrash: homeTrash,
		UID:       os.Getuid(),
		PerMount:  true,
	}

	got, err := resolver.Resolve(testFile)
	if err != nil {
		t.Fatalf("Resolve(/dev/shm/...) error = %v", err)
	}
	if got.Path == "" {
		t.Fatal("expected non-empty root path")
	}
	// Cleanup any created per-mount trash
	if got.Path != homeTrash {
		defer func() { _ = os.RemoveAll(got.Path) }()
	}
}
