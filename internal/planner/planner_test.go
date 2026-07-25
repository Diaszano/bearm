package planner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
)

func TestPlanClassifiesDanglingSymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "missing"), link); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{link},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if len(plan.Targets) != 1 || plan.Targets[0].Kind != domain.TargetSymlink {
		t.Fatalf("Targets = %#v", plan.Targets)
	}
}

func TestPlanUsesFastPathForRecursiveDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "node_modules")
	if err := os.MkdirAll(filepath.Join(directory, "pkg"), 0o700); err != nil {
		t.Fatal(err)
	}

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
		Options: domain.RemoveOptions{
			Recursive:   true,
			Interactive: domain.InteractiveNever,
		},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if plan.Targets[0].RequiresWalk {
		t.Fatal("RequiresWalk = true, want false")
	}
}

func TestPlanRequiresWalkForInteractiveDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "dir")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(directory, "file.txt")
	if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
		Options: domain.RemoveOptions{
			Recursive:   true,
			Interactive: domain.InteractiveAlways,
		},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if len(plan.Targets) != 2 {
		t.Fatalf("len(Targets) = %d, want 2", len(plan.Targets))
	}
	if plan.Targets[0].AbsolutePath != file {
		t.Errorf("Target 0 = %s, want %s", plan.Targets[0].AbsolutePath, file)
	}
	if plan.Targets[1].AbsolutePath != directory {
		t.Errorf("Target 1 = %s, want %s", plan.Targets[1].AbsolutePath, directory)
	}
}

func TestPlanIgnoresMissingOperandUnderForce(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{filepath.Join(t.TempDir(), "missing")},
		Options: domain.RemoveOptions{
			Force:       true,
			Interactive: domain.InteractiveOnce,
		},
	})
	if len(plan.Targets) != 0 || len(failures) != 0 {
		t.Fatalf("plan = %#v, failures = %#v", plan, failures)
	}
	if len(plan.ValidatedOperands) != 0 {
		t.Fatalf("ValidatedOperands = %#v", plan.ValidatedOperands)
	}
	if removal.NeedsOncePrompt(plan) {
		t.Fatal("NeedsOncePrompt() = true, want false for force-missing operand")
	}
}

func TestPlanKeepsDuplicateValidOperandsForOncePrompt(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileBSD,
		Operands: []string{path, path, path, path},
		Options:  domain.RemoveOptions{Interactive: domain.InteractiveOnce},
	})
	if len(failures) != 0 {
		t.Fatalf("failures = %#v", failures)
	}
	if len(plan.Targets) != 1 {
		t.Fatalf("len(Targets) = %d, want 1", len(plan.Targets))
	}
	if !removal.NeedsOncePrompt(plan) {
		t.Fatal("NeedsOncePrompt() = false, want true for four valid operands")
	}
}

func TestPlanExcludesOperandWhoseExpansionFailsFromOncePrompt(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "unreadable")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "file.txt"), []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(directory, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(directory, 0o700); err != nil {
			t.Errorf("restore permissions: %v", err)
		}
	})

	policy, err := safety.NewPolicy(safety.Config{InspectDescendants: true})
	if err != nil {
		t.Fatal(err)
	}
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileBSD,
		Operands: []string{directory},
		Options: domain.RemoveOptions{
			Interactive: domain.InteractiveOnce,
			Recursive:   true,
		},
	})
	if len(failures) != 1 {
		t.Fatalf("failures = %#v, want one expansion failure", failures)
	}
	if len(plan.ValidatedOperands) != 0 {
		t.Fatalf("ValidatedOperands = %#v, want none", plan.ValidatedOperands)
	}
	if removal.NeedsOncePrompt(plan) {
		t.Fatal("NeedsOncePrompt() = true, want false after expansion failure")
	}
}

func TestPlanRejectsDirectoryWithoutRecursiveOrDirectoryOption(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()
	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	_, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{directory},
	})
	if len(failures) != 1 || failures[0].Status != domain.ItemFailed {
		t.Fatalf("failures = %#v", failures)
	}
}
