// Package cli parses Bearm native and rm-compatible command lines.
package cli

import (
	"errors"
	"path/filepath"
)

// InvocationMode identifies the parser selected for an invocation.
type InvocationMode string

const (
	// ModeCompatibility selects rm-compatible parsing.
	ModeCompatibility InvocationMode = "compatibility"
	// ModeNative selects Bearm native subcommands.
	ModeNative InvocationMode = "native"
)

// Invocation contains the mode and arguments after executable/subcommand removal.
type Invocation struct {
	Program string
	Mode    InvocationMode
	Args    []string
}

// ResolveInvocation determines whether argv requests compatibility or native mode.
func ResolveInvocation(argv []string) (Invocation, error) {
	if len(argv) == 0 {
		return Invocation{}, errors.New("argv must include the executable name")
	}

	program := filepath.Base(argv[0])
	if program == "rm" {
		return Invocation{
			Program: program,
			Mode:    ModeCompatibility,
			Args:    append([]string(nil), argv[1:]...),
		}, nil
	}

	if len(argv) > 1 && argv[1] == "rm" {
		return Invocation{
			Program: "rm",
			Mode:    ModeCompatibility,
			Args:    append([]string(nil), argv[2:]...),
		}, nil
	}

	return Invocation{
		Program: program,
		Mode:    ModeNative,
		Args:    append([]string(nil), argv[1:]...),
	}, nil
}
