//go:build linux

package platform

import (
	"path/filepath"

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

// MountPoint returns the top directory on the same device as path.
func MountPoint(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	var stat unix.Stat_t
	if err := unix.Lstat(absolute, &stat); err != nil {
		return "", err
	} else if (stat.Mode & unix.S_IFMT) != unix.S_IFDIR {
		absolute = filepath.Dir(absolute)
	}

	device, err := DeviceID(absolute)
	if err != nil {
		return "", err
	}

	current := filepath.Clean(absolute)
	for {
		parent := filepath.Dir(current)
		if parent == current {
			return current, nil
		}

		parentDevice, err := DeviceID(parent)
		if err != nil {
			return "", err
		}
		if parentDevice != device {
			return current, nil
		}

		current = parent
	}
}
