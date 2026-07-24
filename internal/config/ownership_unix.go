//go:build linux || darwin

package config

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

// CheckSecureFile verifies that a config file is regular, non-symlink, user-owned, and not writable by group or others.
func CheckSecureFile(path string, uid int) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("configuration path must be a regular non-symlink file")
	}
	if info.Mode().Perm()&0o022 != 0 {
		return fmt.Errorf("configuration file permissions %o are insecure", info.Mode().Perm())
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("configuration ownership is unavailable")
	}
	if int(stat.Uid) != uid {
		return fmt.Errorf("configuration file owner %d does not match current user %d", stat.Uid, uid)
	}
	return nil
}
