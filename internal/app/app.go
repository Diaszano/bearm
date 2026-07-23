// Package app coordinates Bearm use cases.
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
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
	return &App{stdin: stdin, out: stdout, err: stderr, info: info}
}

// Run executes one Bearm invocation and returns a process exit code.
func (a *App) Run(_ context.Context, argv []string) int {
	invocation, err := cli.ResolveInvocation(argv)
	if err != nil {
		fmt.Fprintln(a.err, err)
		return 2
	}

	if invocation.Mode == cli.ModeCompatibility {
		return a.runCompatibility(invocation.Args)
	}

	return a.runNative(invocation.Args)
}

func (a *App) runCompatibility(args []string) int {
	profile := resolveProfile(os.Getenv, runtime.GOOS)
	request, err := cli.ParseCompatibility(args, profile)
	if err != nil {
		fmt.Fprintf(a.err, "rm: %v\n", err)
		return usageCode(profile)
	}

	if request.Options.ShowVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}
	if request.Options.ShowHelp {
		fmt.Fprintln(a.out, "Usage: rm [OPTION]... [FILE]...")
		return 0
	}

	if err := request.Validate(); err != nil {
		catalog := i18n.NewCatalog(i18n.ResolveCompatibilityLanguage(os.Getenv))
		fmt.Fprintf(a.err, "rm: %s\n", catalog.Text(i18n.MessageMissingOperand))
		return usageCode(profile)
	}

	return 0
}

func (a *App) runNative(args []string) int {
	catalog := i18n.NewCatalog(i18n.ResolveNativeLanguage(os.Getenv))
	request, err := cli.ParseNative(args)
	if err != nil {
		if strings.Contains(err.Error(), "unknown") || strings.Contains(err.Error(), "missing native") {
			fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
		} else {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
		}
		return 2
	}

	if request.Command == cli.CommandVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}

	fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
	return 2
}

func resolveProfile(getenv func(string) string, goos string) domain.CompatibilityProfile {
	switch strings.ToLower(getenv("BEARM_COMPAT")) {
	case "gnu":
		return domain.ProfileGNU
	case "bsd":
		return domain.ProfileBSD
	case "posix":
		return domain.ProfilePOSIX
	}

	if goos == "darwin" {
		return domain.ProfileBSD
	}
	return domain.ProfileGNU
}

func usageCode(profile domain.CompatibilityProfile) int {
	if profile == domain.ProfileBSD {
		return 64
	}
	return 1
}
