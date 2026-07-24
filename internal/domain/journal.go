package domain

import (
	"errors"
	"time"
)

// JournalAction identifies a durable lifecycle event.
type JournalAction string

const (
	// JournalTrashed records a completed move to trash.
	JournalTrashed JournalAction = "trashed"
	// JournalRestored records a completed restore.
	JournalRestored JournalAction = "restored"
	// JournalPurged records a completed permanent purge.
	JournalPurged JournalAction = "purged"
)

// JournalEvent is one append-only lifecycle event.
type JournalEvent struct {
	SchemaVersion int           `json:"schema_version"`
	Action        JournalAction `json:"action"`
	OccurredAt    time.Time     `json:"occurred_at"`
	Record        TrashRecord   `json:"record"`
}

// Validate validates journal invariants.
func (e JournalEvent) Validate() error {
	if e.SchemaVersion != 1 {
		return errors.New("unsupported journal schema version")
	}
	switch e.Action {
	case JournalTrashed, JournalRestored, JournalPurged:
	default:
		return errors.New("unsupported journal action")
	}
	if e.Record.ItemID == "" || e.Record.OperationID == "" {
		return errors.New("journal event requires item and operation IDs")
	}
	if e.OccurredAt.IsZero() {
		return errors.New("journal event requires occurrence time")
	}
	return nil
}
