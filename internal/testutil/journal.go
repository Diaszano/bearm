package testutil

import (
	"context"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Journal is an in-memory journal test double.
type Journal struct {
	Records []domain.TrashRecord
	Events  []domain.JournalEvent
	Err     error
}

// Append stores trash events.
func (j *Journal) Append(_ context.Context, records []domain.TrashRecord) error {
	j.Records = append(j.Records, records...)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	for _, record := range records {
		j.Events = append(j.Events, record.Event(domain.JournalTrashed, now))
	}
	return j.Err
}

// AppendEvents stores lifecycle events.
func (j *Journal) AppendEvents(_ context.Context, events []domain.JournalEvent) error {
	j.Events = append(j.Events, events...)
	return j.Err
}
