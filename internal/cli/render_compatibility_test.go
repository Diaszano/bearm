package cli_test

import (
	"errors"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
)

func TestGNURendererMissingOperand(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	if got := renderer.MissingOperand(); got != "rm: missing operand\n" {
		t.Fatalf("MissingOperand() = %q", got)
	}
}

func TestBSDRendererUsage(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")
	want := "usage: rm [-f | -i] [-dIRrv] file ...\n"
	if got := renderer.Usage(); got != want {
		t.Fatalf("Usage() = %q, want %q", got, want)
	}
}

func TestGNURendererUnrecognizedLongOption(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	got := renderer.UnsupportedOption(&cli.UsageError{
		Kind: "unsupported-option", Option: "--unknown",
	})
	want := "rm: unrecognized option '--unknown'\n"
	if got != want {
		t.Fatalf("UnsupportedOption() = %q, want %q", got, want)
	}
}

func TestBSDRendererIllegalShortOption(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")
	got := renderer.UnsupportedOption(&cli.UsageError{
		Kind: "unsupported-option", Option: "-P",
	})
	want := "rm: illegal option -- P\n"
	if got != want {
		t.Fatalf("UnsupportedOption() = %q, want %q", got, want)
	}
}

func TestRendererPathError(t *testing.T) {
	t.Parallel()

	renderer := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	got := renderer.PathError("missing", errors.New("No such file or directory"))
	if got != "rm: cannot remove 'missing': No such file or directory\n" {
		t.Fatalf("PathError() = %q", got)
	}
}

func TestRendererPromptsAndEdgeCases(t *testing.T) {
	t.Parallel()

	// GNU / PTBR
	rendererPT := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguagePTBR, "rm")
	if got := rendererPT.MissingOperand(); got != "rm: operando ausente\n" {
		t.Fatalf("MissingOperand() = %q", got)
	}
	if got := rendererPT.Usage(); got != "Usage: rm [OPTION]... [FILE]...\n" {
		t.Fatalf("Usage() = %q", got)
	}

	// BSD PathError
	rendererBSD := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")
	if got := rendererBSD.PathError("missing", errors.New("No such file")); got != "rm: missing: No such file\n" {
		t.Fatalf("PathError() = %q", got)
	}

	// UnsupportedOption with short GNU option
	rendererGNU := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	gotGNU := rendererGNU.UnsupportedOption(&cli.UsageError{
		Kind: "unsupported-option", Option: "-x",
	})
	if gotGNU != "rm: invalid option -- 'x'\n" {
		t.Fatalf("UnsupportedOption() = %q", gotGNU)
	}

	// UnsupportedOption nil check
	if got := rendererGNU.UnsupportedOption(nil); got != "" {
		t.Fatalf("UnsupportedOption(nil) = %q", got)
	}

	// PromptOnce and PromptTarget
	if got := rendererGNU.PromptOnce(); got != "rm: remove all arguments? " {
		t.Fatalf("PromptOnce() = %q", got)
	}
	if got := rendererGNU.PromptTarget("path.txt"); got != "rm: remove 'path.txt'? " {
		t.Fatalf("PromptTarget() = %q", got)
	}
}
