package cli

import (
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
)

// ParseCompatibility parses GNU, BSD, or POSIX rm-compatible arguments.
func ParseCompatibility(args []string, profile domain.CompatibilityProfile) (domain.RemoveRequest, error) {
	state := compatibilityState{
		profile: profile,
		request: domain.RemoveRequest{
			Profile: profile,
			Options: domain.RemoveOptions{
				Interactive:  domain.InteractiveDefault,
				PreserveRoot: domain.PreserveRootDefault,
			},
		},
		parseOptions: true,
	}

	for _, arg := range args {
		if arg == "--" && state.parseOptions {
			state.parseOptions = false
			continue
		}

		if state.parseOptions && isOption(arg) {
			if err := state.applyOption(arg); err != nil {
				return domain.RemoveRequest{}, err
			}
			continue
		}

		state.request.Operands = append(state.request.Operands, arg)
		if profile == domain.ProfileBSD || profile == domain.ProfilePOSIX {
			state.parseOptions = false
		}
	}

	return state.request, nil
}

type compatibilityState struct {
	profile      domain.CompatibilityProfile
	request      domain.RemoveRequest
	parseOptions bool
}

func isOption(arg string) bool {
	return len(arg) > 1 && arg[0] == '-'
}

func (s *compatibilityState) applyOption(arg string) error {
	if strings.HasPrefix(arg, "--") {
		return s.applyLongOption(arg)
	}

	for _, option := range arg[1:] {
		if err := s.applyShortOption("-" + string(option)); err != nil {
			return err
		}
	}

	return nil
}

func (s *compatibilityState) applyShortOption(option string) error {
	switch option {
	case "-f":
		s.request.Options.Force = true
		s.request.Options.Interactive = domain.InteractiveNever
	case "-i":
		s.request.Options.Force = false
		s.request.Options.Interactive = domain.InteractiveAlways
	case "-I":
		s.request.Options.Force = false
		s.request.Options.Interactive = domain.InteractiveOnce
	case "-r", "-R":
		s.request.Options.Recursive = true
	case "-d":
		s.request.Options.Directory = true
	case "-v":
		s.request.Options.Verbose = true
	default:
		return &UsageError{Kind: "unsupported-option", Option: option}
	}

	return nil
}

func (s *compatibilityState) applyLongOption(option string) error {
	switch option {
	case "--force":
		return s.applyShortOption("-f")
	case "--interactive":
		return s.applyInteractive("always")
	case "--recursive", "--Recursive":
		s.request.Options.Recursive = true
	case "--dir", "--directory":
		s.request.Options.Directory = true
	case "--verbose":
		s.request.Options.Verbose = true
	case "--one-file-system":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.OneFileSystem = true
	case "--preserve-root":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootDefault
	case "--preserve-root=all":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootAll
	case "--no-preserve-root":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.PreserveRoot = domain.PreserveRootNone
	case "--help":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.ShowHelp = true
	case "--version":
		if s.profile != domain.ProfileGNU {
			return &UsageError{Kind: "unsupported-option", Option: option}
		}
		s.request.Options.ShowVersion = true
	default:
		if strings.HasPrefix(option, "--interactive=") {
			return s.applyInteractive(strings.TrimPrefix(option, "--interactive="))
		}
		return &UsageError{Kind: "unsupported-option", Option: option}
	}

	return nil
}

func (s *compatibilityState) applyInteractive(value string) error {
	s.request.Options.Force = false

	switch value {
	case "never":
		s.request.Options.Interactive = domain.InteractiveNever
	case "once":
		s.request.Options.Interactive = domain.InteractiveOnce
	case "always":
		s.request.Options.Interactive = domain.InteractiveAlways
	default:
		return &UsageError{Kind: "invalid-interactive", Option: "--interactive", Value: value}
	}

	return nil
}
