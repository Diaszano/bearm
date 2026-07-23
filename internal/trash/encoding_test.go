package trash_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/trash"
)

func TestEncodePathPreservesSafeCharacters(t *testing.T) {
	t.Parallel()

	input := "/home/dias/project/file-name_1.txt"
	if got := trash.EncodePath(input); got != input {
		t.Fatalf("EncodePath() = %q", got)
	}
}

func TestEncodePathEscapesWhitespaceAndPercent(t *testing.T) {
	t.Parallel()

	input := "/home/dias/a b%/ç.txt"
	want := "/home/dias/a%20b%25/%C3%A7.txt"
	if got := trash.EncodePath(input); got != want {
		t.Fatalf("EncodePath() = %q, want %q", got, want)
	}
}

func TestEncodePathUsesUppercaseHex(t *testing.T) {
	t.Parallel()

	if got := trash.EncodePath("\n"); got != "%0A" {
		t.Fatalf("EncodePath() = %q", got)
	}
}
