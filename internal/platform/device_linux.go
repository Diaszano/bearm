//go:build linux

package platform

import (
	"golang.org/x/sys/unix"
)

// DeviceID returns the filesystem device ID for path without following a final symlink.
func DeviceID(path string) (uint64, error) {
	var stat unix.Stat_t
	if err := unix.Lstat(path, &stat); err != nil {
		return 0, err
	}

	return stat.Dev, nil
}
