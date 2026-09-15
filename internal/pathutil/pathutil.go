// Package pathutil provides shared path comparison utilities.
package pathutil

import (
	"path/filepath"
	"strings"
)

// IsEqualOrDescendant reports whether path equals root or is a descendant of root.
// Both path and root are cleaned before comparison.
func IsEqualOrDescendant(path, root string) bool {
	cleanPath := filepath.Clean(path)
	cleanRoot := filepath.Clean(root)
	if cleanPath == cleanRoot {
		return true
	}
	prefix := cleanRoot
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(cleanPath, prefix)
}
