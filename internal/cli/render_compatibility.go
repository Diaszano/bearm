package cli

import (
	"fmt"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
)

// CompatibilityRenderer renders profile-specific rm diagnostics.
type CompatibilityRenderer struct {
	profile  domain.CompatibilityProfile
	language i18n.Language
	program  string
}

// NewCompatibilityRenderer creates a deterministic compatibility renderer.
func NewCompatibilityRenderer(
	profile domain.CompatibilityProfile,
	language i18n.Language,
	program string,
) CompatibilityRenderer {
	return CompatibilityRenderer{
		profile:  profile,
		language: language,
		program:  program,
	}
}

// Usage renders the selected profile usage text.
func (r CompatibilityRenderer) Usage() string {
	if r.profile == domain.ProfileBSD {
		return fmt.Sprintf("usage: %s [-f | -i] [-dIRrv] file ...\n", r.program)
	}
	return fmt.Sprintf("Usage: %s [OPTION]... [FILE]...\n", r.program)
}

// MissingOperand renders a missing operand diagnostic.
func (r CompatibilityRenderer) MissingOperand() string {
	if r.language == i18n.LanguagePTBR {
		return fmt.Sprintf("%s: operando ausente\n", r.program)
	}
	return fmt.Sprintf("%s: missing operand\n", r.program)
}

// UnsupportedOption renders a typed parser usage error.
func (r CompatibilityRenderer) UnsupportedOption(err *UsageError) string {
	if err == nil {
		return ""
	}
	if r.profile == domain.ProfileBSD && strings.HasPrefix(err.Option, "-") && !strings.HasPrefix(err.Option, "--") {
		return fmt.Sprintf("%s: illegal option -- %s\n", r.program, strings.TrimPrefix(err.Option, "-"))
	}
	if strings.HasPrefix(err.Option, "--") {
		return fmt.Sprintf("%s: unrecognized option '%s'\n", r.program, err.Option)
	}
	return fmt.Sprintf("%s: invalid option -- '%s'\n", r.program, strings.TrimPrefix(err.Option, "-"))
}

// PathError renders an operational path error.
func (r CompatibilityRenderer) PathError(path string, err error) string {
	if r.profile == domain.ProfileBSD {
		return fmt.Sprintf("%s: %s: %v\n", r.program, path, err)
	}
	return fmt.Sprintf("%s: cannot remove '%s': %v\n", r.program, path, err)
}

// PromptOnce renders the once-only confirmation prompt.
func (r CompatibilityRenderer) PromptOnce() string {
	return fmt.Sprintf("%s: remove all arguments? ", r.program)
}

// PromptTarget renders a per-target confirmation prompt.
func (r CompatibilityRenderer) PromptTarget(path string) string {
	return fmt.Sprintf("%s: remove '%s'? ", r.program, path)
}
