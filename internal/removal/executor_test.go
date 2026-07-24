package removal_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestExecuteMovesIndependentTargetsAndAppendsJournal(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Interactive: domain.InteractiveNever,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
			{InputPath: "b", AbsolutePath: "/work/b", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}
	if len(backend.Moved) != 2 || len(journal.Records) != 2 {
		t.Fatalf("moved = %#v, journal = %#v", backend.Moved, journal.Records)
	}
}

func TestExecuteContinuesAfterIndependentFailure(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{FailForPath: "/work/a"}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID:      "operation-1",
		Request: domain.RemoveRequest{Profile: domain.ProfileGNU},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
			{InputPath: "b", AbsolutePath: "/work/b", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
	if len(backend.Moved) != 1 || backend.Moved[0] != "/work/b" {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}

func TestExecuteDeclinesOncePromptWithoutMoves(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader("no\n"), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Interactive: domain.InteractiveOnce,
				Recursive:   true,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "dir", AbsolutePath: "/work/dir", Kind: domain.TargetDir},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}

func TestExecuteReportsJournalFailure(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{Err: errors.New("journal failed")}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID:      "operation-1",
		Request: domain.RemoveRequest{Profile: domain.ProfileGNU},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
}

func TestExecuteAcceptsNilVerboseWriter(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		nil, // nil verbose writer!
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Verbose: true,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}
}

func TestExecuteInteractiveAlways(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	// Provide "y\nn\n" - first is accepted, second is declined.
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader("y\nn\n"), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
			Options: domain.RemoveOptions{
				Interactive: domain.InteractiveAlways,
			},
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/a", Kind: domain.TargetFile},
			{InputPath: "b", AbsolutePath: "/work/b", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if result.HasFailures() {
		t.Fatalf("result = %#v", result)
	}

	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d", len(result.Items))
	}
	if result.Items[0].Status != domain.ItemTrashed {
		t.Errorf("Item 0 status = %v", result.Items[0].Status)
	}
	if result.Items[1].Status != domain.ItemDeclined {
		t.Errorf("Item 1 status = %v", result.Items[1].Status)
	}
	if len(backend.Moved) != 1 || backend.Moved[0] != "/work/a" {
		t.Fatalf("moved = %#v", backend.Moved)
	}
}

func TestExecuteSafetyPolicyRejection(t *testing.T) {
	t.Parallel()

	// Configure safety policy to protect /work/protected
	policy, _ := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{"/work/protected"},
	})
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var output bytes.Buffer
	executor := removal.NewExecutor(
		backend,
		journal,
		policy,
		removal.NewPrompter(strings.NewReader(""), &output),
		&output,
	)

	plan := domain.RemovalPlan{
		ID: "operation-1",
		Request: domain.RemoveRequest{
			Profile: domain.ProfileGNU,
		},
		Targets: []domain.PlannedTarget{
			{InputPath: "a", AbsolutePath: "/work/protected", Kind: domain.TargetFile},
		},
	}

	result := executor.Execute(context.Background(), plan)
	if !result.HasFailures() {
		t.Fatal("expected safety failure, got none")
	}
	if result.Items[0].Status != domain.ItemFailed {
		t.Fatalf("status = %v, want failed", result.Items[0].Status)
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("moved = %#v, want 0", backend.Moved)
	}
}
