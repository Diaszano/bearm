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
	"github.com/Diaszano/bearm/internal/platform"
	"runtime"
)

func setupTestJournal(t *testing.T) (*journal.Repository, string, string, string) {
	t.Helper()
	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	if err := os.MkdirAll(filepath.Dir(trashed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(original), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	repoPath := filepath.Join(root, "journal.jsonl")
	repo := journal.New(repoPath, time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-123",
		OperationID:   "op-456",
		OriginalPath:  original,
		TrashedPath:   trashed,
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repo.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}
	return repo, original, trashed, root
}

func TestRunPurge_LastWithYes(t *testing.T) {
	t.Parallel()

	repo, _, trashed, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "purge", "--last", "--yes"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(trashed); !os.IsNotExist(err) {
		t.Fatalf("expected trashed file to be purged, got err: %v", err)
	}
}

func TestRunPurge_InteractiveConfirm(t *testing.T) {
	t.Parallel()

	repo, _, trashed, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader("y\n"),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "purge", "--last"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(trashed); !os.IsNotExist(err) {
		t.Fatalf("expected trashed file to be purged, got err: %v", err)
	}
}

func TestRunPurge_InteractiveDecline(t *testing.T) {
	t.Parallel()

	repo, _, trashed, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader("n\n"),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "purge", "--last"})
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if _, err := os.Stat(trashed); err != nil {
		t.Fatalf("expected file to remain, got: %v", err)
	}
}

func TestRunPurge_ByOperation(t *testing.T) {
	t.Parallel()

	repo, _, trashed, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "purge", "--operation", "op-456", "--yes"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(trashed); !os.IsNotExist(err) {
		t.Fatalf("expected trashed file to be purged, got err: %v", err)
	}
}

func TestRunPurge_OperationNotFound(t *testing.T) {
	t.Parallel()

	repo, _, _, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "purge", "--operation", "non-existent", "--yes"})
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "não encontrada") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunDoctor_Clean(t *testing.T) {
	t.Parallel()

	repo, _, _, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "doctor"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Nenhum problema encontrado") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunDoctor_JSON(t *testing.T) {
	t.Parallel()

	repo, _, _, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "doctor", "--json"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[]") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunDoctor_WithFindings(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repoPath := filepath.Join(root, "journal.jsonl")
	repo := journal.New(repoPath, time.Now)
	// Missing file in trashed path
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-corrupt",
		OperationID:   "op-corrupt",
		OriginalPath:  "/tmp/original",
		TrashedPath:   filepath.Join(root, "non-existent-trash-file"),
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repo.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "doctor"})
	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "item-corrupt") && !strings.Contains(stdout.String(), "non-existent-trash-file") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunList_Tabular(t *testing.T) {
	t.Parallel()

	repo, original, _, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "list"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "item-123") || !strings.Contains(out, original) {
		t.Fatalf("stdout = %q", out)
	}
}

func TestRunRestore_ByOperation(t *testing.T) {
	t.Parallel()

	repo, original, _, _ := setupTestJournal(t)
	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo},
	)

	code := instance.Run(context.Background(), []string{"bearm", "restore", "--operation", "op-456"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(original); err != nil {
		t.Fatalf("expected restored file: %v", err)
	}
}

func TestRunRestore_CollisionOverwritePolicyRejected(t *testing.T) {
	t.Parallel()

	repo, _, _, _ := setupTestJournal(t)
	cfg := config.Default()
	cfg.Restore.CollisionPolicy = "overwrite"

	var stdout, stderr bytes.Buffer
	instance := NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		Dependencies{Repository: repo, Config: cfg},
	)

	code := instance.Run(context.Background(), []string{"bearm", "restore", "--last"})
	if code != 2 {
		t.Fatalf("Run() code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "overwrite exige confirmação explícita") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestResolveConfiguredProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		configured string
		env        string
		goos       string
		want       domain.CompatibilityProfile
	}{
		{"gnu", "", "linux", domain.ProfileGNU},
		{"bsd", "", "linux", domain.ProfileBSD},
		{"posix", "", "linux", domain.ProfilePOSIX},
		{"auto", "gnu", "darwin", domain.ProfileGNU},
		{"auto", "bsd", "linux", domain.ProfileBSD},
		{"auto", "posix", "linux", domain.ProfilePOSIX},
		{"auto", "", "darwin", domain.ProfileBSD},
		{"auto", "", "linux", domain.ProfileGNU},
		{"", "", "darwin", domain.ProfileBSD},
		{"", "", "linux", domain.ProfileGNU},
	}

	for _, tt := range tests {
		getenv := func(key string) string {
			if key == "BEARM_COMPAT" {
				return tt.env
			}
			return ""
		}
		got := resolveConfiguredProfile(tt.configured, getenv, tt.goos)
		if got != tt.want {
			t.Errorf("resolveConfiguredProfile(%q, %q, %q) = %v, want %v", tt.configured, tt.env, tt.goos, got, tt.want)
		}
	}
}

func TestUsageCode(t *testing.T) {
	t.Parallel()

	if got := usageCode(domain.ProfileBSD); got != 64 {
		t.Errorf("usageCode(ProfileBSD) = %d, want 64", got)
	}
	if got := usageCode(domain.ProfileGNU); got != 1 {
		t.Errorf("usageCode(ProfileGNU) = %d, want 1", got)
	}
	if got := usageCode(domain.ProfilePOSIX); got != 1 {
		t.Errorf("usageCode(ProfilePOSIX) = %d, want 1", got)
	}
}

func TestNewConfiguredBackend(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dirs, err := platform.ResolveDirs(func(string) string { return "" }, dir, runtime.GOOS)
	if err != nil {
		t.Fatalf("ResolveDirs() error = %v", err)
	}
	cfg := config.Default()
	cfg.Trash.CustomPath = filepath.Join(dir, "custom-trash")

	backend, err := NewConfiguredBackend(dir, dirs, cfg)
	if err != nil {
		t.Fatalf("NewConfiguredBackend() error = %v", err)
	}
	if backend == nil {
		t.Fatal("expected non-nil backend")
	}

	cfgDefault := config.Default()
	platBackend, err := NewConfiguredBackend(dir, dirs, cfgDefault)
	if err != nil {
		t.Fatalf("NewConfiguredBackend() default error = %v", err)
	}
	if platBackend == nil {
		t.Fatal("expected non-nil platform backend")
	}
}
