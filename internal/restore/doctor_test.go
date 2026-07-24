package restore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/restore"
)

func TestDoctorReportsMissingTrashedPath(t *testing.T) {
	t.Parallel()

	repository := journal.New(filepath.Join(t.TempDir(), "journal.jsonl"), time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file",
		TrashedPath:   filepath.Join(t.TempDir(), "missing"),
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	doctor := restore.NewDoctor(repository)
	findings, err := doctor.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Code != "missing-trash-item" {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestDoctorReportsIncompleteTrailingLine(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	if err := os.WriteFile(path, []byte(`{"schema_version":1`), 0o600); err != nil {
		t.Fatal(err)
	}

	doctor := restore.NewDoctor(journal.New(path, time.Now))
	findings, err := doctor.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Code != "incomplete-journal-line" {
		t.Fatalf("findings = %#v", findings)
	}
}
