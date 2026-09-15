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
	default:
		return fmt.Sprintf("unsupported option %q", e.Option)
	}
}
