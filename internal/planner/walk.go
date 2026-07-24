package planner

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/Diaszano/bearm/internal/platform"
)

// WalkDepthFirst returns children before parents and never follows symlinks.
func WalkDepthFirst(root string, rootDevice uint64, oneFileSystem bool) ([]string, error) {
	if oneFileSystem && rootDevice == 0 {
		dev, err := platform.DeviceID(root)
		if err != nil {
			return nil, err
		}
		rootDevice = dev
	}
	paths := make([]string, 0)

	var visit func(string) error
	visit = func(path string) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}

		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			if oneFileSystem {
				deviceID, err := platform.DeviceID(path)
				if err != nil {
					return err
				}
				if rootDevice != 0 && deviceID != rootDevice {
					return nil
				}
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].Name() < entries[j].Name()
			})
			for _, entry := range entries {
				if err := visit(filepath.Join(path, entry.Name())); err != nil {
					return err
				}
			}
		}

		paths = append(paths, path)
		return nil
	}

	if err := visit(root); err != nil {
		return nil, err
	}
	return paths, nil
}
