package domain_test

import (
	"errors"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestRemovalResultExitCodeSuccess(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{{Path: "file", Status: domain.ItemTrashed}},
	}
	if got := result.ExitCode(domain.ProfileGNU); got != 0 {
		t.Fatalf("ExitCode() = %d", got)
	}
}

func TestRemovalResultExitCodeFailure(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{{Path: "file", Status: domain.ItemFailed, Err: errors.New("failed")}},
	}
	if got := result.ExitCode(domain.ProfileGNU); got != 1 {
		t.Fatalf("GNU ExitCode() = %d", got)
	}
	if got := result.ExitCode(domain.ProfileBSD); got != 1 {
		t.Fatalf("BSD operational ExitCode() = %d", got)
	}
}

func TestRemovalResultHasFailures(t *testing.T) {
	t.Parallel()

	result := domain.RemovalResult{
		Items: []domain.ItemResult{
			{Path: "a", Status: domain.ItemTrashed},
			{Path: "b", Status: domain.ItemFailed, Err: errors.New("failed")},
		},
	}
	if !result.HasFailures() {
		t.Fatal("HasFailures() = false, want true")
	}
}
