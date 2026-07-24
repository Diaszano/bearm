package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := New(strings.NewReader(""), &stdout, &stderr, buildinfo.Info{
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

func TestRunRMForceWithoutOperandsReturnsSuccess(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())

	code := instance.Run(context.Background(), []string{"rm", "-f"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
}

func TestRunRMMissingOperandReturnsGNUFailure(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())
	instance.getenv = func(key string) string {
		switch key {
		case "BEARM_COMPAT":
			return "gnu"
		case "LC_ALL":
			return "C"
		default:
			return ""
		}
	}

	code := instance.Run(context.Background(), []string{"rm"})
	if code != 1 {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "missing operand") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunNativeUnknownCommandUsesPortuguese(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := New(strings.NewReader(""), &stdout, &stderr, buildinfo.Current())
	instance.getenv = func(key string) string {
		if key == "BEARM_LANG" {
			return "pt-BR"
		}
		return ""
	}

	code := instance.Run(context.Background(), []string{"bearm", "unknown"})
	if code != 2 {
		t.Fatalf("Run() code = %d", code)
	}
	if !strings.Contains(stderr.String(), "comando desconhecido") { //nolint:misspell // Portuguese translation for command
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunCompatibilityMovesPlannedTarget(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	source := filepath.Join(root, "file.txt")
	if err := os.WriteFile(source, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	backend := &testutil.Backend{}
	journal := &testutil.Journal{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Backend: backend, Journal: journal, Policy: policy},
	)

	code := instance.Run(context.Background(), []string{"rm", source})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if len(backend.Moved) != 1 || backend.Moved[0] != source {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}
