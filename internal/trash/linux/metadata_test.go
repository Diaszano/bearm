//go:build linux

package linux_test

import (
	"testing"
	"time"

	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func TestRenderTrashInfo(t *testing.T) {
	t.Parallel()

	deletedAt := time.Date(2026, 7, 23, 9, 30, 15, 0, time.FixedZone("BRT", -3*60*60))
	got := string(linuxtrash.RenderTrashInfo("/home/dias/a b.txt", deletedAt))
	want := "[Trash Info]\nPath=/home/dias/a%20b.txt\nDeletionDate=2026-07-23T09:30:15\n"

	if got != want {
		t.Fatalf("RenderTrashInfo() = %q, want %q", got, want)
	}
}
