//go:build linux

package trash

import "golang.org/x/sys/unix"

// RenameNoReplace moves src to dst atomically, failing if dst already exists.
// It uses the RENAME_NOREPLACE flag available since Linux 3.15.
func RenameNoReplace(src, dst string) error {
	return unix.Renameat2(unix.AT_FDCWD, src, unix.AT_FDCWD, dst, unix.RENAME_NOREPLACE)
}
