//go:build darwin

package trash

import "golang.org/x/sys/unix"

// RenameNoReplace moves src to dst atomically, failing if dst already exists.
// It uses renamex_np with RENAME_EXCL on macOS.
func RenameNoReplace(src, dst string) error {
	return unix.RenameatxNp(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_EXCL)
}
