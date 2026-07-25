package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
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

func TestRunBSDInteractiveOnceCountsDuplicateOperands(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{})
	if err != nil {
		t.Fatal(err)
	}
	backend := &testutil.Backend{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader("n\n"),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Backend: backend, Journal: &testutil.Journal{}, Policy: policy},
	)
	instance.getenv = func(key string) string {
		if key == "BEARM_COMPAT" {
			return "bsd"
		}
		return ""
	}

	code := instance.Run(context.Background(), []string{"rm", "-I", path, path, path, path})
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if got := stderr.String(); got != "remove 4 files? " {
		t.Fatalf("stderr = %q", got)
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
}

func TestRunBSDInteractiveOnceFormatsRecursivePrompts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T, root string) []string
		want  func(root string) string
	}{
		{
			name: "one directory",
			setup: func(t *testing.T, root string) []string {
				t.Helper()
				directory := filepath.Join(root, "dir")
				if err := os.Mkdir(directory, 0o700); err != nil {
					t.Fatal(err)
				}
				return []string{directory}
			},
			want: func(root string) string { return "recursively remove " + filepath.Join(root, "dir") + "? " },
		},
		{
			name: "two directories",
			setup: func(t *testing.T, root string) []string {
				t.Helper()
				for _, name := range []string{"one", "two"} {
					if err := os.Mkdir(filepath.Join(root, name), 0o700); err != nil {
						t.Fatal(err)
					}
				}
				return []string{filepath.Join(root, "one"), filepath.Join(root, "two")}
			},
			want: func(string) string { return "recursively remove 2 dirs? " },
		},
		{
			name: "directory and file",
			setup: func(t *testing.T, root string) []string {
				t.Helper()
				directory := filepath.Join(root, "dir")
				if err := os.Mkdir(directory, 0o700); err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(root, "file.txt")
				if err := os.WriteFile(file, []byte("data"), 0o600); err != nil {
					t.Fatal(err)
				}
				return []string{directory, file}
			},
			want: func(root string) string {
				return "recursively remove " + filepath.Join(root, "dir") + " and 1 file? "
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			operands := tt.setup(t, root)
			policy, err := safety.NewPolicy(safety.Config{})
			if err != nil {
				t.Fatal(err)
			}
			backend := &testutil.Backend{}
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			instance := NewWithDependencies(
				strings.NewReader("n\n"), &stdout, &stderr, buildinfo.Current(),
				Dependencies{Backend: backend, Journal: &testutil.Journal{}, Policy: policy},
			)
			instance.getenv = func(key string) string {
				if key == "BEARM_COMPAT" {
					return "bsd"
				}
				return ""
			}

			args := append([]string{"rm", "-Ir"}, operands...)
			code := instance.Run(context.Background(), args)
			if code != 1 {
				t.Fatalf("Run() code = %d, want 1", code)
			}
			if got := stderr.String(); got != tt.want(root) {
				t.Fatalf("stderr = %q, want %q", got, tt.want(root))
			}
			if len(backend.Moved) != 0 {
				t.Fatalf("Moved = %#v", backend.Moved)
			}
		})
	}
}

func TestRunNativeListJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file",
		TrashedPath:   "/trash/file",
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repository},
	)

	code := instance.Run(context.Background(), []string{"bearm", "list", "--json"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"item_id":"item-1"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunNativeRestoreLast(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file")
	original := filepath.Join(root, "work", "file")
	if err := os.MkdirAll(filepath.Dir(trashed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := journal.New(filepath.Join(root, "journal.jsonl"), time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repository},
	)

	code := instance.Run(context.Background(), []string{"bearm", "restore", "--last"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(original); err != nil {
		t.Fatalf("restored path missing: %v", err)
	}
}

func TestRunConfigPath(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{
			Config:     config.Default(),
			ConfigPath: "/home/dias/.config/bearm/config.toml",
		},
	)

	code := instance.Run(context.Background(), []string{"bearm", "config", "path"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "/home/dias/.config/bearm/config.toml" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunConfigCheck(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{
			Config:     config.Default(),
			ConfigPath: "/tmp/config.toml",
		},
	)

	code := instance.Run(context.Background(), []string{"bearm", "config", "check"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Configuração válida") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}
