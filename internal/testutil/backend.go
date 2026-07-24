// Package testutil contains reusable Bearm test doubles.
package testutil

import (
	"context"
	"errors"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Backend is a configurable trash backend test double.
type Backend struct {
	Moved       []string
	FailForPath string
}

// Name returns the fake backend name.
func (b *Backend) Name() string {
	return "fake"
}

// Resolve returns a deterministic fake destination.
func (b *Backend) Resolve(_ context.Context, target domain.PlannedTarget) (domain.Destination, error) {
	if target.AbsolutePath == b.FailForPath {
		return domain.Destination{}, errors.New("resolve failed")
	}
	return domain.Destination{Root: "/trash"}, nil
}

// Move records a fake successful move.
func (b *Backend) Move(
	_ context.Context,
	target domain.PlannedTarget,
	_ domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if target.AbsolutePath == b.FailForPath {
		return domain.TrashRecord{}, errors.New("move failed")
	}
	b.Moved = append(b.Moved, target.AbsolutePath)
	return domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-" + target.InputPath,
		OperationID:   operationID,
		OriginalPath:  target.AbsolutePath,
		TrashedPath:   "/trash/" + target.InputPath,
		Backend:       b.Name(),
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
		Status:        "trashed",
	}, nil
}
