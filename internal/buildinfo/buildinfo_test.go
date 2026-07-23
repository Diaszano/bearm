package buildinfo_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/buildinfo"
)

func TestInfoString(t *testing.T) {
	t.Parallel()

	info := buildinfo.Info{
		Version: "1.2.3",
		Commit:  "abcdef0",
		Date:    "2026-07-23T12:00:00Z",
	}

	want := "Bearm 1.2.3 (abcdef0, 2026-07-23T12:00:00Z)"
	if got := info.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestCurrentUsesDevelopmentDefaults(t *testing.T) {
	t.Parallel()

	info := buildinfo.Current()
	if info.Version == "" || info.Commit == "" || info.Date == "" {
		t.Fatalf("Current() = %#v, expected non-empty fields", info)
	}
}
