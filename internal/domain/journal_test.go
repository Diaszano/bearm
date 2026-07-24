package domain_test

import (
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestTrashRecordEventCreatesVersionedEvent(t *testing.T) {
	t.Parallel()

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file.txt",
		TrashedPath:   "/trash/file.txt",
		Backend:       "fake",
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
		Status:        "trashed",
	}
	occurredAt := time.Date(2026, 7, 23, 12, 0, 1, 0, time.UTC)

	event := record.Event(domain.JournalTrashed, occurredAt)
	if event.SchemaVersion != 1 || event.Action != domain.JournalTrashed {
		t.Fatalf("event = %#v", event)
	}
	if event.Record.ItemID != "item-1" || !event.OccurredAt.Equal(occurredAt) {
		t.Fatalf("event = %#v", event)
	}
}

func TestJournalEventValidateRejectsUnknownAction(t *testing.T) {
	t.Parallel()

	event := domain.JournalEvent{
		SchemaVersion: 1,
		Action:        domain.JournalAction("unknown"),
		OccurredAt:    time.Now(),
		Record:        domain.TrashRecord{ItemID: "item-1", OperationID: "operation-1"},
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestJournalEventValidateRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	event := domain.JournalEvent{
		SchemaVersion: 1,
		Action:        domain.JournalTrashed,
		OccurredAt:    time.Now(),
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
