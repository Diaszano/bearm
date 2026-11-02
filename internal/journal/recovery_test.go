package journal_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
)

func TestReadAllIgnoresIncompleteTrailingLine(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	content := `{"schema_version":1,"action":"trashed","occurred_at":"2026-07-23T12:00:00Z","record":{"schema_version":1,"item_id":"item-1","operation_id":"operation-1","original_path":"/work/a","trashed_path":"/trash/a","backend":"fake","device_id":1,"deleted_at":"2026-07-23T12:00:00Z","status":"trashed"}}` + "\n" +
		`{"schema_version":1,"action":"trashed"`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := journal.New(path, time.Now)
	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(result.Events) != 1 || !result.IncompleteTrailingLine {
		t.Fatalf("result = %#v", result)
	}
}

func TestAppendRejectsIncompleteTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	content := `{"schema_version":1,"action":"trashed","occurred_at":"2026-07-23T12:00:00Z","record":{"schema_version":1,"item_id":"item-1","operation_id":"operation-1","original_path":"/work/a","trashed_path":"/trash/a","backend":"fake","device_id":1,"deleted_at":"2026-07-23T12:00:00Z","status":"trashed"}}` + "\n" + `{"schema_version":1`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	repository := journal.New(path, time.Now)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	err = repository.AppendEvents(context.Background(), []domain.JournalEvent{validEvent()})
	if err == nil {
		t.Fatal("AppendEvents() succeeded with incomplete trailing line")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("append modified journal bytes")
	}
	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != 1 || !result.IncompleteTrailingLine {
		t.Fatalf("result = %#v", result)
	}
}

func TestAppendRejectsValidJSONWithoutNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	if err := os.WriteFile(path, []byte(validEventJSON()), 0o600); err != nil {
		t.Fatal(err)
	}
	repository := journal.New(path, time.Now)
	if err := repository.AppendEvents(context.Background(), []domain.JournalEvent{validEvent()}); err == nil {
		t.Fatal("append succeeded without final newline")
	}
}

func TestAppendAfterCompleteTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	for i := 0; i < 2; i++ {
		if err := repository.AppendEvents(context.Background(), []domain.JournalEvent{validEvent()}); err != nil {
			t.Fatal(err)
		}
	}
	result, err := repository.ReadAll(context.Background())
	if err != nil || len(result.Events) != 2 {
		t.Fatalf("ReadAll() = %d, %v", len(result.Events), err)
	}
}

func validEvent() domain.JournalEvent {
	return domain.TrashRecord{ItemID: "item", OperationID: "operation", DeletedAt: time.Now()}.Event(domain.JournalTrashed, time.Now())
}

func validEventJSON() string {
	data, _ := json.Marshal(validEvent())
	return string(data)
}

func TestAppendInvalidEventDoesNotWrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	err := repository.AppendEvents(context.Background(), []domain.JournalEvent{{SchemaVersion: 0}})
	if err == nil {
		t.Fatal("AppendEvents() accepted invalid event")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("journal created: %v", err)
	}
}

func TestAppendConcurrentCompleteEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	const writers = 8
	var wg sync.WaitGroup
	errs := make(chan error, writers)
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- repository.AppendEvents(context.Background(), []domain.JournalEvent{validEvent()})
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	result, err := repository.ReadAll(context.Background())
	if err != nil || len(result.Events) != writers {
		t.Fatalf("ReadAll() = %d, %v", len(result.Events), err)
	}
}
