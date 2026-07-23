package cli_test

import (
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
)

func TestResolveInvocationUsesRMModeFromArgvZero(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"/usr/local/bin/rm", "-rf", "build"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeCompatibility {
		t.Fatalf("Mode = %q", got.Mode)
	}
	want := []string{"-rf", "build"}
	if !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("Args = %#v, want %#v", got.Args, want)
	}
}

func TestResolveInvocationUsesRMSubcommand(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"bearm", "rm", "-f", "missing"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeCompatibility {
		t.Fatalf("Mode = %q", got.Mode)
	}
	want := []string{"-f", "missing"}
	if !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("Args = %#v, want %#v", got.Args, want)
	}
}

func TestResolveInvocationUsesNativeMode(t *testing.T) {
	t.Parallel()

	got, err := cli.ResolveInvocation([]string{"bearm", "list", "--limit", "10"})
	if err != nil {
		t.Fatalf("ResolveInvocation() error = %v", err)
	}

	if got.Mode != cli.ModeNative {
		t.Fatalf("Mode = %q", got.Mode)
	}
}

func TestResolveInvocationRejectsEmptyArgv(t *testing.T) {
	t.Parallel()

	if _, err := cli.ResolveInvocation(nil); err == nil {
		t.Fatal("ResolveInvocation() error = nil, want non-nil")
	}
}
