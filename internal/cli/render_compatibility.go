package cli

import (
	"errors"
	"fmt"
	"os"
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
		if r.profile == domain.ProfileGNU {
			return fmt.Sprintf("%s: operando ausente\nExperimente '%s --help' para mais informações.\n", r.program, r.program)
		}
		return fmt.Sprintf("%s: operando ausente\n", r.program)
	}
	if r.profile == domain.ProfileGNU {
		return fmt.Sprintf("%s: missing operand\nTry '%s --help' for more information.\n", r.program, r.program)
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
		return fmt.Sprintf("%s: %s: %s\n", r.program, path, formatError(err))
	}
	return fmt.Sprintf("%s: cannot remove '%s': %s\n", r.program, path, formatError(err))
}

func formatError(err error) string {
	if err == nil {
		return ""
	}
	var pe *os.PathError
	msg := err.Error()
	if errors.As(err, &pe) && pe.Err != nil {
		msg = pe.Err.Error()
	}
	if len(msg) > 0 && msg[0] >= 'a' && msg[0] <= 'z' {
		msg = string(msg[0]-'a'+'A') + msg[1:]
	}
	return msg
}

// PromptOnce renders the once-only confirmation prompt.
func (r CompatibilityRenderer) PromptOnce() string {
	return fmt.Sprintf("%s: remove all arguments? ", r.program)
}

// PromptTarget renders a per-target confirmation prompt.
func (r CompatibilityRenderer) PromptTarget(path string) string {
	return fmt.Sprintf("%s: remove '%s'? ", r.program, path)
}
