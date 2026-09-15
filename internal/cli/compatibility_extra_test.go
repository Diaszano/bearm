package cli_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
)

func TestParseCompatibilityLongOptionsGNU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		validate func(t *testing.T, req domain.RemoveRequest)
	}{
		{
			name: "--force",
			args: []string{"--force", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.Force {
					t.Error("expected Force = true")
				}
			},
		},
		{
			name: "--interactive",
			args: []string{"--interactive", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.Interactive != domain.InteractiveAlways {
					t.Errorf("expected InteractiveAlways, got %v", req.Options.Interactive)
				}
			},
		},
		{
			name: "--recursive and --Recursive",
			args: []string{"--recursive", "--Recursive", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.Recursive {
					t.Error("expected Recursive = true")
				}
			},
		},
		{
			name: "--dir and --directory",
			args: []string{"--dir", "--directory", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.Directory {
					t.Error("expected Directory = true")
				}
			},
		},
		{
			name: "--verbose",
			args: []string{"--verbose", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.Verbose {
					t.Error("expected Verbose = true")
				}
			},
		},
		{
			name: "--one-file-system",
			args: []string{"--one-file-system", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.OneFileSystem {
					t.Error("expected OneFileSystem = true")
				}
			},
		},
		{
			name: "--preserve-root",
			args: []string{"--preserve-root", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.PreserveRoot != domain.PreserveRootDefault {
					t.Errorf("expected PreserveRootDefault, got %v", req.Options.PreserveRoot)
				}
			},
		},
		{
			name: "--preserve-root=all",
			args: []string{"--preserve-root=all", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.PreserveRoot != domain.PreserveRootAll {
					t.Errorf("expected PreserveRootAll, got %v", req.Options.PreserveRoot)
				}
			},
		},
		{
			name: "--no-preserve-root",
			args: []string{"--no-preserve-root", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.PreserveRoot != domain.PreserveRootNone {
					t.Errorf("expected PreserveRootNone, got %v", req.Options.PreserveRoot)
				}
			},
		},
		{
			name: "--help",
			args: []string{"--help"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.ShowHelp {
					t.Error("expected ShowHelp = true")
				}
			},
		},
		{
			name: "--version",
			args: []string{"--version"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if !req.Options.ShowVersion {
					t.Error("expected ShowVersion = true")
				}
			},
		},
		{
			name: "--interactive=never",
			args: []string{"--interactive=never", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.Interactive != domain.InteractiveNever {
					t.Errorf("expected InteractiveNever, got %v", req.Options.Interactive)
				}
			},
		},
		{
			name: "--interactive=once",
			args: []string{"--interactive=once", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.Interactive != domain.InteractiveOnce {
					t.Errorf("expected InteractiveOnce, got %v", req.Options.Interactive)
				}
			},
		},
		{
			name: "--interactive=always",
			args: []string{"--interactive=always", "f"},
			validate: func(t *testing.T, req domain.RemoveRequest) {
				if req.Options.Interactive != domain.InteractiveAlways {
					t.Errorf("expected InteractiveAlways, got %v", req.Options.Interactive)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cli.ParseCompatibility(tt.args, domain.ProfileGNU)
			if err != nil {
				t.Fatalf("ParseCompatibility(%v) error = %v", tt.args, err)
			}
			tt.validate(t, got)
		})
	}
}

func TestParseCompatibilityLongOptionsNonGNU(t *testing.T) {
	t.Parallel()

	unsupported := []string{
		"--one-file-system",
		"--preserve-root",
		"--preserve-root=all",
		"--no-preserve-root",
		"--help",
		"--version",
	}

	for _, opt := range unsupported {
		opt := opt
		t.Run(opt, func(t *testing.T) {
			t.Parallel()
			_, err := cli.ParseCompatibility([]string{opt, "file"}, domain.ProfileBSD)
			if err == nil {
				t.Fatalf("expected error for option %s on BSD profile, got nil", opt)
			}
			var usageErr *cli.UsageError
			if !errors.As(err, &usageErr) {
				t.Errorf("expected UsageError, got %T: %v", err, err)
			}
		})
	}
}

func TestUsageError_Error(t *testing.T) {
	t.Parallel()

	err := &cli.UsageError{Kind: "unsupported-option", Option: "--invalid"}
	msg := err.Error()
	if !strings.Contains(msg, "--invalid") {
		t.Errorf("err.Error() = %q, want it to contain option name", msg)
	}
}

func TestRendererPromptOnce(t *testing.T) {
	t.Parallel()

	rendererGNU := cli.NewCompatibilityRenderer(domain.ProfileGNU, i18n.LanguageEN, "rm")
	planGNU := domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Recursive: false},
		},
		ValidatedOperands: []domain.ValidatedOperand{
			{InputPath: "a", Kind: domain.TargetFile},
		},
	}
	if got := rendererGNU.PromptOnce(planGNU); got != "rm: remove all arguments? " {
		t.Errorf("PromptOnce GNU = %q, want 'rm: remove all arguments? '", got)
	}

	rendererBSD := cli.NewCompatibilityRenderer(domain.ProfileBSD, i18n.LanguageEN, "rm")

	// Non-recursive BSD
	planBSDNonRec := domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Recursive: false},
		},
		ValidatedOperands: []domain.ValidatedOperand{
			{InputPath: "a", Kind: domain.TargetFile},
			{InputPath: "b", Kind: domain.TargetFile},
		},
	}
	if got := rendererBSD.PromptOnce(planBSDNonRec); got != "remove 2 files? " {
		t.Errorf("PromptOnce BSD non-rec = %q, want 'remove 2 files? '", got)
	}

	// Recursive BSD with 1 directory
	planBSD1Dir := domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Recursive: true},
		},
		ValidatedOperands: []domain.ValidatedOperand{
			{InputPath: "mydir", Kind: domain.TargetDir},
		},
	}
	if got := rendererBSD.PromptOnce(planBSD1Dir); got != "recursively remove mydir? " {
		t.Errorf("PromptOnce BSD 1 dir = %q, want 'recursively remove mydir? '", got)
	}

	// Recursive BSD with 2 directories and 2 files
	planBSD2Dirs2Files := domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Recursive: true},
		},
		ValidatedOperands: []domain.ValidatedOperand{
			{InputPath: "dir1", Kind: domain.TargetDir},
			{InputPath: "dir2", Kind: domain.TargetDir},
			{InputPath: "file1", Kind: domain.TargetFile},
			{InputPath: "file2", Kind: domain.TargetFile},
		},
	}
	if got := rendererBSD.PromptOnce(planBSD2Dirs2Files); got != "recursively remove 2 dirs and 2 files? " {
		t.Errorf("PromptOnce BSD 2 dirs 2 files = %q, want 'recursively remove 2 dirs and 2 files? '", got)
	}

	// Recursive BSD with 1 dir and 1 file
	planBSD1Dir1File := domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Recursive: true},
		},
		ValidatedOperands: []domain.ValidatedOperand{
			{InputPath: "dir1", Kind: domain.TargetDir},
			{InputPath: "file1", Kind: domain.TargetFile},
		},
	}
	if got := rendererBSD.PromptOnce(planBSD1Dir1File); got != "recursively remove dir1 and 1 file? " {
		t.Errorf("PromptOnce BSD 1 dir 1 file = %q, want 'recursively remove dir1 and 1 file? '", got)
	}

	// PromptTarget
	if got := rendererBSD.PromptTarget("/path/to/target"); got != "rm: remove '/path/to/target'? " {
		t.Errorf("PromptTarget = %q, want 'rm: remove \\'/path/to/target\\'? '", got)
	}
}
