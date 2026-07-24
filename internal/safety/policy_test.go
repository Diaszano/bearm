package safety_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func TestPolicyRejectsRootAndDotSegments(t *testing.T) {
	t.Parallel()

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{"/", ".", "..", "/tmp/.", "/tmp/.."} {
		if err := policy.Check(path); err == nil {
			t.Errorf("Check(%q) error = nil, want non-nil", path)
		}
	}
}

func TestPolicyRejectsInternalRoots(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{
			filepath.Join(root, "trash"),
			filepath.Join(root, "state"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(root, "trash"),
		filepath.Join(root, "trash", "item"),
		filepath.Join(root, "state", "journal.jsonl"),
	} {
		if err := policy.Check(path); err == nil {
			t.Errorf("Check(%q) error = nil, want non-nil", path)
		}
	}
}

func TestPolicyEnforcesAllowedRoots(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	allowed := filepath.Join(root, "allowed")
	policy, err := safety.NewPolicy(safety.Config{AllowedRoots: []string{allowed}})
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Check(filepath.Join(allowed, "file")); err != nil {
		t.Fatalf("allowed Check() error = %v", err)
	}
	if err := policy.Check(filepath.Join(root, "outside")); err == nil {
		t.Fatal("outside Check() error = nil, want non-nil")
	}
}

func TestPolicyDoesNotResolveFinalSymlinkLexically(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policy, err := safety.NewPolicy(safety.Config{AllowedRoots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, "link")
	if err := policy.Check(path); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestPolicyInspectDescendants(t *testing.T) {
	t.Parallel()

	policy, err := safety.NewPolicy(safety.Config{InspectDescendants: true})
	if err != nil {
		t.Fatal(err)
	}

	if !policy.InspectDescendants() {
		t.Fatal("InspectDescendants() = false, want true")
	}
}
