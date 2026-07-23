package cli

import "fmt"

// UsageError reports invalid command-line syntax without rendering text.
type UsageError struct {
	Kind   string
	Option string
	Value  string
}

func (e *UsageError) Error() string {
	switch e.Kind {
	case "invalid-interactive":
		return fmt.Sprintf("invalid interactive value %q", e.Value)
	case "missing-value":
		return fmt.Sprintf("missing value for %q", e.Option)
	default:
		return fmt.Sprintf("unsupported option %q", e.Option)
	}
}
