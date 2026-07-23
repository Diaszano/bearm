// Package app coordinates Bearm use cases.
package app

import (
	"context"
	"fmt"
	"io"

	"github.com/Diaszano/bearm/internal/buildinfo"
)

// App is the Bearm application shell.
type App struct {
	stdin io.Reader
	out   io.Writer
	err   io.Writer
	info  buildinfo.Info
}

// New creates an application with explicit input and output streams.
func New(stdin io.Reader, stdout, stderr io.Writer, info buildinfo.Info) *App {
	return &App{
		stdin: stdin,
		out:   stdout,
		err:   stderr,
		info:  info,
	}
}

// Run executes one Bearm invocation and returns a process exit code.
func (a *App) Run(_ context.Context, argv []string) int {
	if len(argv) == 2 && argv[1] == "version" {
		_, _ = fmt.Fprintln(a.out, a.info.String())
		return 0
	}

	_, _ = fmt.Fprintln(a.err, "bearm: comando desconhecido") //nolint:misspell // Portuguese translation for command
	return 2
}
