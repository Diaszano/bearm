package integration_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestCompatibilityCannotRemoveHardProtectedRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	protected := filepath.Join(root, "state")
	if err := os.MkdirAll(protected, 0o700); err != nil {
		t.Fatal(err)
	}

	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{protected},
	})
	if err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: &testutil.Backend{},
			Journal: &testutil.Journal{},
			Policy:  policy,
			Config:  config.Default(),
		},
	)

	code := instance.Run(context.Background(), []string{"rm", "-rf", protected})
	if code == 0 {
		t.Fatalf("Run() code = 0, stderr = %q", stderr.String())
	}
	if _, err := os.Stat(protected); err != nil {
		t.Fatalf("protected root changed: %v", err)
	}
}

func TestCompatibilityCannotRemoveSystemRootWithNoPreserveRoot(t *testing.T) {
	t.Parallel()

	policy, _ := safety.NewPolicy(safety.Config{})
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: &testutil.Backend{},
			Journal: &testutil.Journal{},
			Policy:  policy,
			Config:  config.Default(),
		},
	)

	code := instance.Run(context.Background(), []string{"rm", "-rf", "--no-preserve-root", "/"})
	if code == 0 {
		t.Fatal("Run() code = 0, want nonzero")
	}
}
