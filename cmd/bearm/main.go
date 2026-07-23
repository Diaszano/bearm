// Package main is the entry point for the Bearm CLI tool.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Diaszano/bearm/internal/app"
	"github.com/Diaszano/bearm/internal/buildinfo"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	instance := app.New(os.Stdin, os.Stdout, os.Stderr, buildinfo.Current())
	return instance.Run(ctx, os.Args)
}
