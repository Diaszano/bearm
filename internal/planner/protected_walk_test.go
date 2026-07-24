package planner_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
	"github.com/Diaszano/bearm/internal/safety"
)

func TestExpandTargetRetainsProtectedDescendantAndAncestors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "keep", "important.txt")
	movable := filepath.Join(root, "remove", "temporary.txt")
	for _, path := range []string{protected, movable} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	matcher, err := safety.CompilePatterns([]string{filepath.ToSlash(protected)})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := safety.NewPolicy(safety.Config{
		Patterns:           matcher,
		InspectDescendants: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	targets, skipped, err := planner.ExpandTarget(
		domain.PlannedTarget{
			InputPath: root, AbsolutePath: root, Kind: domain.TargetDir,
		},
		domain.RemoveOptions{Recursive: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}

	paths := map[string]bool{}
	for _, target := range targets {
		paths[target.AbsolutePath] = true
	}
	if paths[protected] || paths[filepath.Dir(protected)] || paths[root] {
		t.Fatalf("protected path or ancestor planned: %#v", paths)
	}
	if !paths[movable] || !paths[filepath.Dir(movable)] {
		t.Fatalf("movable subtree missing: %#v", paths)
	}
	if len(skipped) != 1 || skipped[0].Path != protected {
		t.Fatalf("skipped = %#v", skipped)
	}
}

func TestExpandTargetVerbosePlansEveryEntry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	file := filepath.Join(root, "sub", "file.txt")
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, _ := safety.NewPolicy(safety.Config{})
	targets, _, err := planner.ExpandTarget(
		domain.PlannedTarget{InputPath: root, AbsolutePath: root, Kind: domain.TargetDir},
		domain.RemoveOptions{Recursive: true, Verbose: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 3 {
		t.Fatalf("targets = %#v, want file, subdirectory, root", targets)
	}
}

func TestProtectedPlanNeverContainsAncestorThatWouldMoveProtectedContent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "keep", "file.txt")
	if err := os.MkdirAll(filepath.Dir(protected), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(protected, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	matcher, _ := safety.CompilePatterns([]string{filepath.ToSlash(protected)})
	policy, _ := safety.NewPolicy(safety.Config{Patterns: matcher, InspectDescendants: true})
	targets, _, err := planner.ExpandTarget(
		domain.PlannedTarget{AbsolutePath: root, Kind: domain.TargetDir},
		domain.RemoveOptions{Recursive: true},
		policy,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range targets {
		relative, err := filepath.Rel(target.AbsolutePath, protected)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Fatalf("planned ancestor %q would move protected %q", target.AbsolutePath, protected)
		}
	}
}
