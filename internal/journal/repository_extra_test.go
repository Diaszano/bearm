package journal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
)

func TestRepository_Path(t *testing.T) {
	path := "/fake/path"
	r := journal.New(path, time.Now)
	if r.Path() != path {
		t.Errorf("Path() = %q, want %q", r.Path(), path)
	}
}

func TestAppendEvents_Empty(t *testing.T) {
	r := journal.New("/fake/path", time.Now)
	if err := r.AppendEvents(context.Background(), nil); err != nil {
		t.Errorf("AppendEvents() error = %v", err)
	}
}

func TestAppendEvents_ContextCanceled(t *testing.T) {
	r := journal.New("/fake/path", time.Now)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.AppendEvents(ctx, []domain.JournalEvent{
		{
			SchemaVersion: 1,
			Action:        domain.JournalTrashed,
			OccurredAt:    time.Now(),
			Record: domain.TrashRecord{
				ItemID:      "i1",
				OperationID: "o1",
			},
		},
	})
	if err == nil {
		t.Errorf("AppendEvents() expected context canceled error")
	}
}

func TestAppendEvents_InvalidEvent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)
	err := r.AppendEvents(context.Background(), []domain.JournalEvent{
		{SchemaVersion: 0}, // Invalid schema
	})
	if err == nil {
		t.Errorf("AppendEvents() expected validation error")
	}
}

func TestReadAll_NotExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-exists.jsonl")
	r := journal.New(path, time.Now)
	result, err := r.ReadAll(context.Background())
	if err != nil {
		t.Errorf("ReadAll() unexpected error: %v", err)
	}
	if len(result.Events) != 0 {
		t.Errorf("ReadAll() expected 0 events")
	}
}

func TestReadAll_ContextCanceled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)
	// Write a valid event first
	if err := r.AppendEvents(context.Background(), []domain.JournalEvent{
		{
			SchemaVersion: 1,
			Action:        domain.JournalTrashed,
			OccurredAt:    time.Now(),
			Record: domain.TrashRecord{
				ItemID:      "i1",
				OperationID: "o1",
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := r.ReadAll(ctx)
	if err == nil {
		t.Errorf("ReadAll() expected context canceled error")
	}
}

func TestReadAll_IncompleteLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)

	// Create corrupted file
	if err := os.WriteFile(path, []byte("{\"schema_version\": 1, \"action\": \"trashed\""), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := r.ReadAll(context.Background())
	if err != nil {
		t.Errorf("ReadAll() unexpected error: %v", err)
	}
	if !result.IncompleteTrailingLine {
		t.Errorf("ReadAll() expected IncompleteTrailingLine to be true")
	}
}

func TestReadAll_CorruptedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)

	// Complete line but invalid JSON
	if err := os.WriteFile(path, []byte("invalid-json\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := r.ReadAll(context.Background())
	if err == nil {
		t.Errorf("ReadAll() expected unmarshal error")
	}
}

func TestActiveItems_Lifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)

	now := time.Now()
	rec := domain.TrashRecord{
		ItemID:      "i1",
		OperationID: "o1",
		DeletedAt:   now,
	}

	// Trash and then Purge
	events := []domain.JournalEvent{
		rec.Event(domain.JournalTrashed, now),
		rec.Event(domain.JournalPurged, now.Add(time.Second)),
	}
	if err := r.AppendEvents(context.Background(), events); err != nil {
		t.Fatal(err)
	}

	active, err := r.ActiveItems(context.Background())
	if err != nil {
		t.Fatalf("ActiveItems() error = %v", err)
	}
	if len(active) != 0 {
		t.Errorf("ActiveItems() expected 0 active items, got %d", len(active))
	}
}

func TestFindItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)

	now := time.Now()
	rec := domain.TrashRecord{
		ItemID:      "i1",
		OperationID: "o1",
		DeletedAt:   now,
	}
	if err := r.AppendEvents(context.Background(), []domain.JournalEvent{
		rec.Event(domain.JournalTrashed, now),
	}); err != nil {
		t.Fatal(err)
	}

	t.Run("found", func(t *testing.T) {
		items, err := r.FindItems(context.Background(), []string{"i1"})
		if err != nil {
			t.Fatalf("FindItems() error = %v", err)
		}
		if len(items) != 1 || items[0].ItemID != "i1" {
			t.Errorf("FindItems() mismatch")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := r.FindItems(context.Background(), []string{"i2"})
		if err == nil {
			t.Errorf("FindItems() expected error for not found item")
		}
	})
}

func TestLatestOperation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	r := journal.New(path, time.Now)

	t.Run("empty", func(t *testing.T) {
		_, err := r.LatestOperation(context.Background())
		if err == nil {
			t.Errorf("LatestOperation() expected error for empty journal")
		}
	})

	t.Run("with records", func(t *testing.T) {
		now := time.Now()
		rec1 := domain.TrashRecord{ItemID: "i1", OperationID: "o1", DeletedAt: now}
		rec2 := domain.TrashRecord{ItemID: "i2", OperationID: "o2", DeletedAt: now.Add(time.Second)}

		if err := r.AppendEvents(context.Background(), []domain.JournalEvent{
			rec1.Event(domain.JournalTrashed, now),
			rec2.Event(domain.JournalTrashed, now.Add(time.Second)),
		}); err != nil {
			t.Fatal(err)
		}

		items, err := r.LatestOperation(context.Background())
		if err != nil {
			t.Fatalf("LatestOperation() error = %v", err)
		}
		// The active items are sorted by DeletedAt descending
		// So "o2" which has a later DeletedAt will be first.
		if len(items) != 1 || items[0].OperationID != "o2" {
			t.Errorf("LatestOperation() expected o2, got %v", items)
		}
	})
}
