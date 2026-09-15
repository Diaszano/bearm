//go:build linux

package trash

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// RenameNoReplace moves src to dst without replacing an existing destination.
// It falls back to a best-effort rename when the filesystem rejects RENAME_NOREPLACE.
func RenameNoReplace(src, dst string) error {
	return renameNoReplace(src, dst, func(src, dst string) error {
		return unix.Renameat2(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_NOREPLACE)
	})
}

func renameNoReplace(src, dst string, rename func(string, string) error) error {
	err := rename(src, dst)
	if err == nil || (!errors.Is(err, unix.EINVAL) && !errors.Is(err, unix.ENOSYS) && !errors.Is(err, unix.EOPNOTSUPP)) {
		return err
	}

	// ponytail: best-effort fallback can race external writers; remove when all supported filesystems implement RENAME_NOREPLACE.
	if _, statErr := os.Lstat(dst); statErr == nil {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: os.ErrExist}
	} else if !os.IsNotExist(statErr) {
		return statErr
	}
	return os.Rename(src, dst)
}
