//go:build !linux && !darwin

package config

import "os"

// CheckSecureFile is a stub for non-Unix platforms.
func CheckSecureFile(path string, uid int) error {
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	// On unsupported/other operating systems, we don't enforce Unix file permissions.
	return nil
}
