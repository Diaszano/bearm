package app_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Info{
		Version: "1.0.0",
		Commit:  "abcdef0",
		Date:    "2026-07-23T12:00:00Z",
	})

	code := instance.Run(context.Background(), []string{"bearm", "version"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if got := stdout.String(); got != "Bearm 1.0.0 (abcdef0, 2026-07-23T12:00:00Z)\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"bearm", "unknown"})
	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "comando desconhecido") { //nolint:misspell // Portuguese translation for command
		t.Fatalf("stderr = %q", stderr.String())
	}
}
