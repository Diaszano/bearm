// Package safety enforces Bearm hard and configured path protections.
package safety

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// Config contains immutable safety policy inputs.
type Config struct {
	HardProtectedRoots []string
	AllowedRoots       []string
}

// Policy validates lexical absolute target paths.
type Policy struct {
	hardProtected []string
	allowed       []string
}

// NewPolicy validates and creates a safety policy.
func NewPolicy(config Config) (*Policy, error) {
	policy := &Policy{}
	var err error

	policy.hardProtected, err = normalizeRoots(config.HardProtectedRoots)
	if err != nil {
		return nil, fmt.Errorf("hard protected roots: %w", err)
	}
	policy.allowed, err = normalizeRoots(config.AllowedRoots)
	if err != nil {
		return nil, fmt.Errorf("allowed roots: %w", err)
	}

	return policy, nil
}

// Check verifies that path is not hard protected and is inside configured scope.
func (p *Policy) Check(path string) error {
	cleaned := filepath.Clean(path)
	if !filepath.IsAbs(cleaned) {
		return errors.New("safety checks require an absolute path")
	}

	if cleaned == string(filepath.Separator) {
		return errors.New("system root may not be removed")
	}
	if filepath.Base(path) == "." || filepath.Base(path) == ".." || path == "." || path == ".." {
		return errors.New("dot and dot-dot may not be removed")
	}

	for _, root := range p.hardProtected {
		if isEqualOrDescendant(cleaned, root) {
			return fmt.Errorf("path is protected by Bearm: %s", root)
		}
	}

	if len(p.allowed) == 0 {
		return nil
	}
	for _, root := range p.allowed {
		if isEqualOrDescendant(cleaned, root) {
			return nil
		}
	}

	return errors.New("path is outside allowed roots")
}

func normalizeRoots(values []string) ([]string, error) {
	roots := make([]string, 0, len(values))
	for _, value := range values {
		if !filepath.IsAbs(value) {
			return nil, fmt.Errorf("%q is not absolute", value)
		}
		roots = append(roots, filepath.Clean(value))
	}
	return roots, nil
}

func isEqualOrDescendant(path, root string) bool {
	if path == root {
		return true
	}
	prefix := root
	if !strings.HasSuffix(prefix, string(filepath.Separator)) {
		prefix += string(filepath.Separator)
	}
	return strings.HasPrefix(path, prefix)
}
