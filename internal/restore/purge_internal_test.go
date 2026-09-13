package restore

import (
	"context"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestIsInsideTrash(t *testing.T) {
	// Empty roots means true
	if !isInsideTrash("/any/path", nil) {
		t.Error("expected true for empty roots")
	}

	roots := []string{"/trash"}
	if !isInsideTrash("/trash/file.txt", roots) {
		t.Error("expected true for file inside root")
	}
	if !isInsideTrash("/trash", roots) {
		t.Error("expected true for root itself")
	}
	if isInsideTrash("/other/file.txt", roots) {
		t.Error("expected false for file outside root")
	}
}

func TestPurgeContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := NewPurger(&dummyJournal{}, time.Now, nil)
	results := p.Purge(ctx, []domain.TrashRecord{{TrashedPath: "/dummy"}}, true)
	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("expected failure due to context cancellation, got: %v", results)
	}
}

func TestPurgeOutsideTrashRoot(t *testing.T) {
	p := NewPurger(&dummyJournal{}, time.Now, []string{"/valid/trash"})
	results := p.Purge(context.Background(), []domain.TrashRecord{{TrashedPath: "/invalid/trash/file"}}, true)
	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("expected failure due to outside trash roots, got: %v", results)
	}
}
