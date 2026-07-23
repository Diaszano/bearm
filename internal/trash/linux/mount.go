//go:build linux

package linux

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/platform"
	"golang.org/x/sys/unix"
)

// Root is one FreeDesktop trash root.
type Root struct {
	Path             string
	MountPoint       string
	RelativeInfoPath bool
	UID              int
	CheckSecurity    bool
}

// Ensure creates and validates the trash skeleton.
func (r Root) Ensure() error {
	if !filepath.IsAbs(r.Path) {
		return errors.New("trash root must be absolute")
	}

	if err := ensureRealDirectory(r.Path, r.UID, r.CheckSecurity); err != nil {
		return err
	}
	if err := ensureRealDirectory(filepath.Join(r.Path, "files"), r.UID, r.CheckSecurity); err != nil {
		return err
	}
	return ensureRealDirectory(filepath.Join(r.Path, "info"), r.UID, r.CheckSecurity)
}

// RootResolver selects the same-filesystem trash root for a target.
type RootResolver struct {
	HomeTrash string
	UID       int
	PerMount  bool
}

// Resolve selects the FreeDesktop trash root for targetPath.
func (r RootResolver) Resolve(targetPath string) (Root, error) {
	if !filepath.IsAbs(r.HomeTrash) {
		return Root{}, errors.New("home trash must be absolute")
	}

	targetDevice, err := platform.DeviceID(targetPath)
	if err != nil {
		return Root{}, err
	}

	homeParent := filepath.Dir(r.HomeTrash)
	if err := os.MkdirAll(homeParent, 0o700); err != nil {
		return Root{}, err
	}
	homeDevice, err := platform.DeviceID(homeParent)
	if err != nil {
		return Root{}, err
	}

	if !r.PerMount || targetDevice == homeDevice {
		root := Root{Path: r.HomeTrash, UID: r.UID, CheckSecurity: false}
		return root, root.Ensure()
	}

	mountPoint, err := platform.MountPoint(targetPath)
	if err != nil {
		return Root{}, err
	}

	adminTrash := filepath.Join(mountPoint, ".Trash")
	if validAdminTrash(adminTrash) {
		root := Root{
			Path:             filepath.Join(adminTrash, fmt.Sprintf("%d", r.UID)),
			MountPoint:       mountPoint,
			RelativeInfoPath: true,
			UID:              r.UID,
			CheckSecurity:    true,
		}
		if err := root.Ensure(); err == nil {
			return root, nil
		}
	}

	root := Root{
		Path:             filepath.Join(mountPoint, fmt.Sprintf(".Trash-%d", r.UID)),
		MountPoint:       mountPoint,
		RelativeInfoPath: true,
		UID:              r.UID,
		CheckSecurity:    true,
	}
	if err := root.Ensure(); err == nil {
		return root, nil
	}

	homeRoot := Root{Path: r.HomeTrash, UID: r.UID, CheckSecurity: false}
	if err := homeRoot.Ensure(); err != nil {
		return Root{}, fmt.Errorf("per-mount trash failed and home trash fallback failed: %w", err)
	}
	return homeRoot, nil
}

func ensureRealDirectory(path string, uid int, checkSecurity bool) error {
	var stat unix.Stat_t
	err := unix.Lstat(path, &stat)
	if err == nil {
		if (stat.Mode & unix.S_IFMT) != unix.S_IFDIR {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		if checkSecurity {
			if int(stat.Uid) != uid {
				return fmt.Errorf("%s is owned by uid %d, expected %d", path, stat.Uid, uid)
			}
			if (stat.Mode & 0777) != 0700 {
				return fmt.Errorf("%s permissions are %o, expected 0700", path, stat.Mode&0777)
			}
		}
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}
	return os.MkdirAll(path, 0700)
}

func validAdminTrash(path string) bool {
	var stat unix.Stat_t
	err := unix.Lstat(path, &stat)
	if err != nil {
		return false
	}
	if (stat.Mode & unix.S_IFMT) != unix.S_IFDIR {
		return false
	}
	return (stat.Mode & unix.S_ISVTX) != 0
}
