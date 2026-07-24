// Package main is the entry point for the Bearm CLI tool.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/safety"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}

	dirs, err := platform.ResolveDirs(os.Getenv, home, runtime.GOOS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bearm: %v\n", err)
		return 3
	}

	configPath := filepath.Join(dirs.ConfigRoot, "config.toml")
	settings, err := config.Load(configPath, os.Getenv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bearm: configuração inválida: %v\n", err)
		return 3
	}

	patterns, err := safety.CompilePatterns(settings.Safety.ProtectedPatterns)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bearm: padrões de proteção inválidos: %v\n", err)
		return 3
	}

	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{dirs.ConfigRoot, dirs.StateRoot, dirs.DataRoot},
		AllowedRoots:       settings.Safety.AllowedRoots,
		Patterns:           patterns,
		InspectDescendants: settings.Safety.InspectDescendants,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}

	backend, err := app.NewConfiguredBackend(home, settings)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 3
	}

	journalRepo := journal.New(filepath.Join(dirs.StateRoot, "journal.jsonl"), time.Now)

	dependencies := app.Dependencies{
		Backend:    backend,
		Journal:    journalRepo,
		Repository: journalRepo,
		Policy:     policy,
		Config:     settings,
		ConfigPath: configPath,
	}

	instance := app.NewWithDependencies(os.Stdin, os.Stdout, os.Stderr, buildinfo.Current(), dependencies)
	return instance.Run(ctx, os.Args)
}
