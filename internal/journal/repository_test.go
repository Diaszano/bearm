package journal_test

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
)

func TestAppendAndReadAll(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	clock := func() time.Time {
		return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	}
	repository := journal.New(path, clock)

	records := []domain.TrashRecord{
		{
			SchemaVersion: 1,
			ItemID:        "item-1",
			OperationID:   "operation-1",
			OriginalPath:  "/work/a",
			TrashedPath:   "/trash/a",
			Backend:       "fake",
			DeletedAt:     clock(),
			Status:        "trashed",
		},
		{
			SchemaVersion: 1,
			ItemID:        "item-2",
			OperationID:   "operation-1",
			OriginalPath:  "/work/b",
			TrashedPath:   "/trash/b",
			Backend:       "fake",
			DeletedAt:     clock(),
			Status:        "trashed",
		},
	}

	if err := repository.Append(context.Background(), records); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(result.Events) != 2 || result.IncompleteTrailingLine {
		t.Fatalf("result = %#v", result)
	}
}

func TestActiveItemsAppliesLifecycleEvents(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	clock := func() time.Time {
		return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	}
	repository := journal.New(path, clock)

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/a",
		TrashedPath:   "/trash/a",
		Backend:       "fake",
		DeletedAt:     clock(),
		Status:        "trashed",
	}
	events := []domain.JournalEvent{
		record.Event(domain.JournalTrashed, clock()),
		record.Event(domain.JournalRestored, clock().Add(time.Second)),
	}

	if err := repository.AppendEvents(context.Background(), events); err != nil {
		t.Fatal(err)
	}

	active, err := repository.ActiveItems(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Fatalf("active = %#v, want empty", active)
	}
}

func TestConcurrentAppendPreservesCompleteLines(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)

	const count = 25
	errs := make(chan error, count)
	for index := 0; index < count; index++ {
		index := index
		go func() {
			record := domain.TrashRecord{
				SchemaVersion: 1,
				ItemID:        fmt.Sprintf("item-%d", index),
				OperationID:   fmt.Sprintf("operation-%d", index),
				OriginalPath:  fmt.Sprintf("/work/%d", index),
				TrashedPath:   fmt.Sprintf("/trash/%d", index),
				Backend:       "fake",
				DeletedAt:     time.Now(),
				Status:        "trashed",
			}
			errs <- repository.Append(context.Background(), []domain.TrashRecord{record})
		}()
	}

	for index := 0; index < count; index++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}

	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != count {
		t.Fatalf("events = %d, want %d", len(result.Events), count)
	}
}
