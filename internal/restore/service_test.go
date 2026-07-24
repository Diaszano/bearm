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

func TestRestoreMovesItemBackAndAppendsEvent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	if err := os.MkdirAll(filepath.Dir(trashed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	journal := &testutil.Journal{}
	service := restore.NewService(journal, func() time.Time {
		return time.Date(2026, 7, 23, 13, 0, 0, 0, time.UTC)
	})

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}

	results := service.Restore(context.Background(), []domain.TrashRecord{record}, restore.CollisionFail)
	if len(results) != 1 || results[0].Status != domain.ItemRestored {
		t.Fatalf("results = %#v", results)
	}
	if got, err := os.ReadFile(original); err != nil || string(got) != "data" {
		t.Fatalf("original data = %q, error = %v", got, err)
	}
	if len(journal.Events) != 1 || journal.Events[0].Action != domain.JournalRestored {
		t.Fatalf("events = %#v", journal.Events)
	}
}

func TestRestoreFailsOnDestinationCollision(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	for _, path := range []string{trashed, original} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	service := restore.NewService(&testutil.Journal{}, time.Now)
	results := service.Restore(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
	}}, restore.CollisionFail)

	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("results = %#v", results)
	}
}

func TestRestoreRenameCreatesUniqueDestination(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	for _, path := range []string{trashed, original} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	service := restore.NewService(&testutil.Journal{}, time.Now)
	results := service.Restore(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
	}}, restore.CollisionRename)

	if len(results) != 1 || results[0].Status != domain.ItemRestored {
		t.Fatalf("results = %#v", results)
	}
	if results[0].Path != original+".restored.1" {
		t.Fatalf("restored path = %q", results[0].Path)
	}
}
