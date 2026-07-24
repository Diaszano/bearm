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

func BenchmarkAppendOperation(b *testing.B) {
	path := filepath.Join(b.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	records := make([]domain.TrashRecord, 100)
	for index := range records {
		records[index] = domain.TrashRecord{
			SchemaVersion: 1,
			ItemID:        fmt.Sprintf("item-%d", index),
			OperationID:   "operation",
			OriginalPath:  fmt.Sprintf("/work/%d", index),
			TrashedPath:   fmt.Sprintf("/trash/%d", index),
			Backend:       "benchmark",
			DeletedAt:     time.Now(),
			Status:        "trashed",
		}
	}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := repository.Append(context.Background(), records); err != nil {
			b.Fatal(err)
		}
	}
}
