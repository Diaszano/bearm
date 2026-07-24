package restore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/restore"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestPurgeRequiresConfirmation(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	purger := restore.NewPurger(&testutil.Journal{}, time.Now)
	results := purger.Purge(context.Background(), []domain.TrashRecord{{
		ItemID: "item-1", OperationID: "operation-1", TrashedPath: path,
	}}, false)

	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("results = %#v", results)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should remain: %v", err)
	}
}

func TestPurgeRemovesTargetAndAppendsEvent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "directory")
	if err := os.MkdirAll(filepath.Join(path, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}

	journal := &testutil.Journal{}
	purger := restore.NewPurger(journal, time.Now)
	results := purger.Purge(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID:        "item-1", OperationID: "operation-1", TrashedPath: path,
	}}, true)

	if len(results) != 1 || results[0].Status != domain.ItemPurged {
		t.Fatalf("results = %#v", results)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("path still exists: %v", err)
	}
	if len(journal.Events) != 1 || journal.Events[0].Action != domain.JournalPurged {
		t.Fatalf("events = %#v", journal.Events)
	}
}
