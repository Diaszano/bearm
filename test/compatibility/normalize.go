package compatibility

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// TreeEntry is one normalized fixture path after execution.
type TreeEntry struct {
	Path string
	Mode fs.FileMode
	Link string
}

// NormalizeTree returns the logical source-tree state and ignores Bearm internals.
func NormalizeTree(root string) ([]TreeEntry, error) {
	entries := make([]TreeEntry, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}

		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == ".bearm-trash" || relative == ".bearm-state" || relative == ".bearm-config" ||
			hasPrefix(relative, ".bearm-trash") || hasPrefix(relative, ".bearm-state") || hasPrefix(relative, ".bearm-config") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		item := TreeEntry{Path: filepath.ToSlash(relative), Mode: info.Mode()}
		if info.Mode()&os.ModeSymlink != 0 {
			item.Link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		entries = append(entries, item)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Path < entries[j].Path
	})
	return entries, nil
}

func hasPrefix(path, root string) bool {
	return path != root && len(path) > len(root) && path[:len(root)] == root &&
		(path[len(root)] == '/' || path[len(root)] == filepath.Separator)
}
