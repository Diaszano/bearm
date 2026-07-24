package planner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
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

	directory := t.TempDir()
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
	if !plan.Targets[0].RequiresWalk {
		t.Fatal("RequiresWalk = false, want true")
	}
}

func TestPlanIgnoresMissingOperandUnderForce(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	instance := planner.New(policy, func() (string, error) { return "operation-1", nil })

	plan, failures := instance.Plan(context.Background(), domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{filepath.Join(t.TempDir(), "missing")},
		Options:  domain.RemoveOptions{Force: true},
	})
	if len(plan.Targets) != 0 || len(failures) != 0 {
		t.Fatalf("plan = %#v, failures = %#v", plan, failures)
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
