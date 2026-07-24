package planner

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/safety"
)

// ExpandTarget expands a traversal-required directory into child-first targets.
func ExpandTarget(
	root domain.PlannedTarget,
	options domain.RemoveOptions,
	policy *safety.Policy,
) ([]domain.PlannedTarget, []domain.ItemResult, error) {
	type entry struct {
		path     string
		kind     domain.TargetKind
		deviceID uint64
	}

	entries := make([]entry, 0)
	protected := make(map[string]struct{})
	skipped := make([]domain.ItemResult, 0)

	var visit func(string) error
	visit = func(path string) error {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		deviceID, err := platform.DeviceID(path)
		if err != nil {
			return err
		}
		if options.OneFileSystem && root.DeviceID != 0 && deviceID != root.DeviceID {
			return nil
		}

		if info.IsDir() && info.Mode()&os.ModeSymlink == 0 {
			children, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			sort.Slice(children, func(i, j int) bool {
				return children[i].Name() < children[j].Name()
			})
			for _, child := range children {
				if err := visit(filepath.Join(path, child.Name())); err != nil {
					return err
				}
			}
		}

		kind := classify(info)
		if err := policy.Check(path); err != nil {
			protected[path] = struct{}{}
			skipped = append(skipped, domain.ItemResult{
				Path: path, Status: domain.ItemSkipped, Err: err,
			})
		}
		entries = append(entries, entry{path: path, kind: kind, deviceID: deviceID})
		return nil
	}

	if err := visit(root.AbsolutePath); err != nil {
		return nil, nil, err
	}

	blocked := make(map[string]struct{})
	for path := range protected {
		current := path
		for isEqualOrDescendant(current, root.AbsolutePath) {
			blocked[current] = struct{}{}
			if current == root.AbsolutePath {
				break
			}
			current = filepath.Dir(current)
		}
	}

	targets := make([]domain.PlannedTarget, 0, len(entries))
	for _, current := range entries {
		if _, isBlocked := blocked[current.path]; isBlocked {
			continue
		}
		inputPath := current.path
		if root.InputPath != "" {
			if rel, err := filepath.Rel(root.AbsolutePath, current.path); err == nil {
				if rel == "." {
					inputPath = root.InputPath
				} else {
					inputPath = filepath.Join(root.InputPath, rel)
				}
			}
		}
		targets = append(targets, domain.PlannedTarget{
			InputPath:    inputPath,
			AbsolutePath: current.path,
			Kind:         current.kind,
			DeviceID:     current.deviceID,
			RequiresWalk: false,
		})
	}
	return targets, skipped, nil
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
