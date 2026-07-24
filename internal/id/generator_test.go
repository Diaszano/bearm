package id_test

import (
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/id"
)

func TestNewReturnsFixedLengthLowercaseHex(t *testing.T) {
	t.Parallel()

	got, err := id.New()
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if len(got) != 32 {
		t.Fatalf("len(New()) = %d, want 32", len(got))
	}
	if got != strings.ToLower(got) {
		t.Fatalf("New() = %q, want lowercase", got)
	}
}

func TestNewReturnsUniqueValues(t *testing.T) {
	t.Parallel()

	first, err := id.New()
	if err != nil {
		t.Fatal(err)
	}
	second, err := id.New()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("duplicate IDs %q", first)
	}
}
