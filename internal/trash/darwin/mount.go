//go:build darwin

package darwin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/platform"
	"golang.org/x/sys/unix"
)

// Root is one macOS trash root.
type Root struct {
	Path string
}

// Ensure creates Bearm metadata storage inside the system-visible trash.
func (r Root) Ensure() error {
	if !filepath.IsAbs(r.Path) {
		return errors.New("trash root must be absolute")
	}
	for _, directory := range []string{
		r.Path,
		filepath.Join(r.Path, ".bearm-info"),
	} {
		var stat unix.Stat_t
		err := unix.Lstat(directory, &stat)
		if err == nil {
			if (stat.Mode & unix.S_IFMT) != unix.S_IFDIR {
				return fmt.Errorf("%s is not a real directory", directory)
			}
			continue
		}
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return err
		}
		if err := os.Chmod(directory, 0o700); err != nil {
			return err
		}
	}
	return nil
}

// RootResolver selects a macOS home or per-volume trash.
type RootResolver struct {
	Home       string
	UID        int
	DeviceID   func(string) (uint64, error)
	MountPoint func(string) (string, error)
}

// NewRootResolver creates a resolver using real platform functions.
func NewRootResolver(home string, uid int) RootResolver {
	return RootResolver{
		Home:       home,
		UID:        uid,
		DeviceID:   platform.DeviceID,
		MountPoint: platform.MountPoint,
	}
}

// Resolve selects a same-filesystem system trash for targetPath.
func (r RootResolver) Resolve(targetPath string) (Root, error) {
	if !filepath.IsAbs(r.Home) {
		return Root{}, errors.New("home directory must be absolute")
	}

	targetDevice, err := r.DeviceID(targetPath)
	if err != nil {
		return Root{}, err
	}
	homeDevice, err := r.DeviceID(r.Home)
	if err != nil {
		return Root{}, err
	}

	if targetDevice == homeDevice {
		root := Root{Path: filepath.Join(r.Home, ".Trash")}
		return root, root.Ensure()
	}

	mountPoint, err := r.MountPoint(targetPath)
	if err != nil {
		return Root{}, err
	}
	root := Root{Path: filepath.Join(mountPoint, ".Trashes", fmt.Sprintf("%d", r.UID))}
	return root, root.Ensure()
}
