package cli_test

import (
	"fmt"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func BenchmarkParseCompatibility(b *testing.B) {
	args := []string{"-rfv", "--one-file-system", "directory", "file.txt"}

	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		if _, err := cli.ParseCompatibility(args, domain.ProfileGNU); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTenThousandOperands(b *testing.B) {
	args := make([]string, 10001)
	args[0] = "-f"
	for index := 1; index < len(args); index++ {
		args[index] = fmt.Sprintf("file-%d", index)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := cli.ParseCompatibility(args, domain.ProfileGNU); err != nil {
			b.Fatal(err)
		}
	}
}
