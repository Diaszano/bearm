package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Purger permanently removes explicitly selected trash items.
type Purger struct {
	journal EventAppender
	clock   func() time.Time
}

// NewPurger creates a permanent purge service.
func NewPurger(journal EventAppender, clock func() time.Time) *Purger {
	return &Purger{journal: journal, clock: clock}
}

// Purge permanently removes records only when confirmed is true.
func (p *Purger) Purge(
	ctx context.Context,
	records []domain.TrashRecord,
	confirmed bool,
) []domain.ItemResult {
	if !confirmed {
		results := make([]domain.ItemResult, 0, len(records))
		for _, record := range records {
			results = append(results, failed(record.TrashedPath, errors.New("purge confirmation required")))
		}
		return results
	}

	results := make([]domain.ItemResult, 0, len(records))
	events := make([]domain.JournalEvent, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			results = append(results, failed(record.TrashedPath, err))
			break
		}
		if err := os.RemoveAll(record.TrashedPath); err != nil {
			results = append(results, failed(record.TrashedPath, err))
			continue
		}

		purged := record
		purged.Status = "purged"
		events = append(events, purged.Event(domain.JournalPurged, p.clock()))
		results = append(results, domain.ItemResult{
			Path: record.TrashedPath, Status: domain.ItemPurged, Record: &purged,
		})
	}

	if len(events) > 0 {
		if err := p.journal.AppendEvents(ctx, events); err != nil {
			results = append(results, failed("", fmt.Errorf("append purge journal: %w", err)))
		}
	}

	return results
}
