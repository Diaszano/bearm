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

	// 4. When checkSecurity is false, verify that it accepts different permissions/owner
	if err := ensureRealDirectory(badPermDir, uid, false); err != nil {
		t.Errorf("expected success with checkSecurity=false, got err: %v", err)
	}
	if err := ensureRealDirectory(okDir, uid+1, false); err != nil {
		t.Errorf("expected success with checkSecurity=false, got err: %v", err)
	}
}
