package trash_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/trash"
)

func TestRenderMetadata(t *testing.T) {
	when := time.Date(2026, 9, 14, 9, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	got, err := trash.RenderMetadata("/tmp/ação\n.txt", when)
	want := "{\"schema_version\":1,\"original_path\":\"/tmp/ação\\n.txt\",\"deleted_at\":\"2026-09-14T12:00:00Z\"}\n"
	if err != nil || string(got) != want {
		t.Fatalf("metadata = %q, %v; want %q", got, err, want)
	}
	_, err = trash.RenderMetadata("/tmp/file", time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected time serialization error")
	}
	if bytes.Count(got, []byte{'\n'}) != 1 {
		t.Fatal("metadata must contain exactly one trailing newline")
	}
}
