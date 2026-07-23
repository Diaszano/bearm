package domain

import (
	"context"
	"time"
)

// Destination is a reserved trash destination.
type Destination struct {
	Root       string
	FilesDir   string
	InfoDir    string
	TargetPath string
	InfoPath   string
}

// TrashRecord describes one completed trash move.
type TrashRecord struct {
	SchemaVersion int       `json:"schema_version"`
	ItemID        string    `json:"item_id"`
	OperationID   string    `json:"operation_id"`
	OriginalPath  string    `json:"original_path"`
	TrashedPath   string    `json:"trashed_path"`
	Backend       string    `json:"backend"`
	DeviceID      uint64    `json:"device_id"`
	DeletedAt     time.Time `json:"deleted_at"`
	Status        string    `json:"status"`
}

// TrashBackend moves planned targets into a platform trash.
type TrashBackend interface {
	// Name returns the name of the trash backend implementation.
	Name() string
	// Resolve determines the destination paths under the trash directory for a planned target.
	Resolve(context.Context, PlannedTarget) (Destination, error)
	// Move moves the planned target into the reserved trash destination and returns a TrashRecord.
	Move(context.Context, PlannedTarget, Destination, string) (TrashRecord, error)
}
