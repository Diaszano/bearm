package journal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
