package cli_test

import (
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
)

func TestParseNativeList(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"list", "--limit", "25", "--json"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if got.Command != cli.CommandList || got.Limit != 25 || !got.JSON {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativeRestoreLast(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"restore", "--last"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if got.Command != cli.CommandRestore || !got.Last {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativePurgeItems(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"purge", "--yes", "item-1", "item-2"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}

	if !got.Yes || !reflect.DeepEqual(got.ItemIDs, []string{"item-1", "item-2"}) {
		t.Fatalf("request = %#v", got)
	}
}

func TestParseNativeRejectsConflictingRestoreSelectors(t *testing.T) {
	t.Parallel()

	_, err := cli.ParseNative([]string{"restore", "--last", "--operation", "op-1"})
	if err == nil {
		t.Fatal("ParseNative() error = nil, want non-nil")
	}
}

func TestParseNativeVersion(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"version"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}
	if got.Command != cli.CommandVersion {
		t.Fatalf("request = %#v", got)
	}

	_, err = cli.ParseNative([]string{"version", "extra"})
	if err == nil {
		t.Fatal("expected error for extra version args")
	}
}

func TestParseNativeDoctor(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseNative([]string{"doctor"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}
	if got.Command != cli.CommandDoctor || got.JSON {
		t.Fatalf("request = %#v", got)
	}

	gotJSON, err := cli.ParseNative([]string{"doctor", "--json"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}
	if gotJSON.Command != cli.CommandDoctor || !gotJSON.JSON {
		t.Fatalf("request = %#v", gotJSON)
	}

	_, err = cli.ParseNative([]string{"doctor", "--invalid"})
	if err == nil {
		t.Fatal("expected error for invalid doctor arg")
	}
}

func TestParseNativeConfig(t *testing.T) {
	t.Parallel()

	gotPath, err := cli.ParseNative([]string{"config", "path"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}
	if gotPath.Command != cli.CommandConfig || gotPath.ConfigOp != "path" {
		t.Fatalf("request = %#v", gotPath)
	}

	gotCheck, err := cli.ParseNative([]string{"config", "check"})
	if err != nil {
		t.Fatalf("ParseNative() error = %v", err)
	}
	if gotCheck.Command != cli.CommandConfig || gotCheck.ConfigOp != "check" {
		t.Fatalf("request = %#v", gotCheck)
	}

	_, err = cli.ParseNative([]string{"config"})
	if err == nil {
		t.Fatal("expected error for missing config subcommand")
	}

	_, err = cli.ParseNative([]string{"config", "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown config subcommand")
	}
}

func TestParseNativeEmptyAndUnknown(t *testing.T) {
	t.Parallel()

	_, err := cli.ParseNative([]string{})
	if err == nil {
		t.Fatal("expected error for empty args")
	}

	_, err = cli.ParseNative([]string{"unknown"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestParseNativeListErrors(t *testing.T) {
	t.Parallel()

	testCases := [][]string{
		{"list", "--operation"},
		{"list", "--limit"},
		{"list", "--limit", "abc"},
		{"list", "--limit", "0"},
		{"list", "--limit", "-5"},
		{"list", "--unsupported"},
	}

	for _, tc := range testCases {
		_, err := cli.ParseNative(tc)
		if err == nil {
			t.Errorf("expected error for args %v, got nil", tc)
		}
	}
}

func TestParseNativeSelectionErrors(t *testing.T) {
	t.Parallel()

	testCases := [][]string{
		{"restore"},                             // no selectors
		{"restore", "--yes", "item-1"},          // --yes not allowed for restore
		{"restore", "--operation"},              // missing operation value
		{"restore", "--invalid-flag"},           // unsupported option starting with -
		{"purge", "--last", "--operation", "o"}, // multiple selectors
		{"purge"},                               // no selectors
	}

	for _, tc := range testCases {
		_, err := cli.ParseNative(tc)
		if err == nil {
			t.Errorf("expected error for args %v, got nil", tc)
		}
	}
}
