package pathutil_test

import (
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/pathutil"
)

func TestNewIDReturnsFixedLengthLowercaseHex(t *testing.T) {
	t.Parallel()

	got, err := pathutil.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("len(NewID()) = %d, want 32", len(got))
	}
	if got != strings.ToLower(got) {
		t.Fatalf("NewID() = %q, want lowercase", got)
	}
}

func TestNewIDReturnsUniqueValues(t *testing.T) {
	t.Parallel()

	first, err := pathutil.NewID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := pathutil.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("duplicate IDs %q", first)
	}
}
