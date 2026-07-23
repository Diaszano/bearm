package cli_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func TestParseGNUCombinedOptions(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-rfv", "build"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Recursive || !got.Options.Force || !got.Options.Verbose {
		t.Fatalf("Options = %#v", got.Options)
	}
	if !reflect.DeepEqual(got.Operands, []string{"build"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseGNUOptionOrderForceThenInteractive(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-f", "-i", "file"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if got.Options.Force {
		t.Fatal("Force = true, want false after later -i")
	}
	if got.Options.Interactive != domain.InteractiveAlways {
		t.Fatalf("Interactive = %q", got.Options.Interactive)
	}
}

func TestParseGNUOptionOrderInteractiveThenForce(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"-i", "-f", "file"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Force {
		t.Fatal("Force = false, want true after later -f")
	}
	if got.Options.Interactive != domain.InteractiveNever {
		t.Fatalf("Interactive = %q", got.Options.Interactive)
	}
}

func TestParseGNUAcceptsOptionsAfterOperand(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"build", "-rf"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !got.Options.Recursive || !got.Options.Force {
		t.Fatalf("Options = %#v", got.Options)
	}
	if !reflect.DeepEqual(got.Operands, []string{"build"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseBSDStopsOptionsAtFirstOperand(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"build", "-rf"}, domain.ProfileBSD)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if got.Options.Recursive || got.Options.Force {
		t.Fatalf("Options = %#v", got.Options)
	}
	want := []string{"build", "-rf"}
	if !reflect.DeepEqual(got.Operands, want) {
		t.Fatalf("Operands = %#v, want %#v", got.Operands, want)
	}
}

func TestParseCompatibilityHonorsDoubleDash(t *testing.T) {
	t.Parallel()

	got, err := cli.ParseCompatibility([]string{"--", "-rf"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}

	if !reflect.DeepEqual(got.Operands, []string{"-rf"}) {
		t.Fatalf("Operands = %#v", got.Operands)
	}
}

func TestParseGNUInteractiveValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		arg  string
		want domain.InteractiveMode
	}{
		{"--interactive=never", domain.InteractiveNever},
		{"--interactive=once", domain.InteractiveOnce},
		{"--interactive=always", domain.InteractiveAlways},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.arg, func(t *testing.T) {
			t.Parallel()

			got, err := cli.ParseCompatibility([]string{tt.arg, "file"}, domain.ProfileGNU)
			if err != nil {
				t.Fatalf("ParseCompatibility() error = %v", err)
			}
			if got.Options.Interactive != tt.want {
				t.Fatalf("Interactive = %q, want %q", got.Options.Interactive, tt.want)
			}
		})
	}
}

func TestParseCompatibilityRejectsUnsupportedOption(t *testing.T) {
	t.Parallel()

	_, err := cli.ParseCompatibility([]string{"-P", "file"}, domain.ProfileBSD)
	var usageErr *cli.UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("error = %T %v, want *cli.UsageError", err, err)
	}
	if usageErr.Option != "-P" {
		t.Fatalf("Option = %q", usageErr.Option)
	}
}

func TestParsedRequestValidationHandlesForceWithoutOperands(t *testing.T) {
	t.Parallel()

	request, err := cli.ParseCompatibility([]string{"-f"}, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}
	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestParsedRequestValidationRejectsMissingOperand(t *testing.T) {
	t.Parallel()

	request, err := cli.ParseCompatibility(nil, domain.ProfileGNU)
	if err != nil {
		t.Fatalf("ParseCompatibility() error = %v", err)
	}
	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
