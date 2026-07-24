package integration_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/safety"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestCancelledContextPreventsNewMoves(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	policy, _ := safety.NewPolicy(safety.Config{})
	backend := &testutil.Backend{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{
			Backend: backend,
			Journal: &testutil.Journal{},
			Policy:  policy,
			Config:  config.Default(),
		},
	)

	code := instance.Run(ctx, []string{"rm", "file"})
	if code == 0 {
		t.Fatal("Run() code = 0, want nonzero")
	}
	if len(backend.Moved) != 0 {
		t.Fatalf("Moved = %#v", backend.Moved)
	}
	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatalf("ctx.Err() = %v", ctx.Err())
	}
}
