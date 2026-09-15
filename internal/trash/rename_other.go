//go:build !linux && !darwin

package trash

import "os"

// RenameNoReplace moves src to dst, failing if dst already exists.
// On platforms without kernel-level RENAME_NOREPLACE support, this
// uses a best-effort Lstat+Rename approach.
func RenameNoReplace(src, dst string) error {
	if _, err := os.Lstat(dst); err == nil {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: os.ErrExist}
	}
	return os.Rename(src, dst)
}
