package restore

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestIsCrossDevice(t *testing.T) {
	// Not cross device
	if isCrossDevice(errors.New("generic error")) {
		t.Error("expected generic error to not be cross device")
	}

	// Test syscall.EXDEV
	if !isCrossDevice(syscall.EXDEV) {
		t.Error("expected syscall.EXDEV to be cross device")
	}

	// Test os.LinkError with EXDEV
	linkErr := &os.LinkError{
		Op:  "rename",
		Old: "old",
		New: "new",
		Err: syscall.EXDEV,
	}
	if !isCrossDevice(linkErr) {
		t.Error("expected os.LinkError with EXDEV to be cross device")
	}
}

type dummyJournal struct {
	err error
}

func (d *dummyJournal) AppendEvents(_ context.Context, _ []domain.JournalEvent) error {
	return d.err
}

func TestRestoreContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := NewService(&dummyJournal{}, time.Now)
	results := svc.Restore(ctx, []domain.TrashRecord{{OriginalPath: "/dummy"}}, CollisionFail)
	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("expected failure due to context cancellation, got: %v", results)
	}
}
