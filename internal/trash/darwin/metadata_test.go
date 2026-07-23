//go:build darwin

package darwin_test

import (
	"strings"
	"testing"
	"time"

	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func TestRenderMetadataIsStableJSON(t *testing.T) {
	t.Parallel()

	metadata := darwintrash.Metadata{
		SchemaVersion: 1,
		OriginalPath:  "/Users/dias/a b.txt",
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}

	got, err := darwintrash.RenderMetadata(metadata)
	if err != nil {
		t.Fatalf("RenderMetadata() error = %v", err)
	}
	if !strings.HasSuffix(string(got), "\n") {
		t.Fatalf("metadata must end with newline: %q", got)
	}
	if !strings.Contains(string(got), `"original_path":"/Users/dias/a b.txt"`) {
		t.Fatalf("metadata = %q", got)
	}
}
