package restore

import (
	"context"
	"os"

	"github.com/Diaszano/bearm/internal/journal"
)

// Finding is one non-destructive doctor diagnostic.
type Finding struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// Check returns diagnostics without mutating journal or trash.
func Check(ctx context.Context, repository *journal.Repository) ([]Finding, error) {
	readResult, err := repository.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0)
	if readResult.IncompleteTrailingLine {
		findings = append(findings, Finding{
			Code:    "incomplete-journal-line",
			Message: "journal contains an incomplete trailing line",
			Path:    repository.Path(),
		})
	}

	active, err := repository.ActiveItems(ctx)
	if err != nil {
		return nil, err
	}
	for _, record := range active {
		if _, err := os.Lstat(record.TrashedPath); os.IsNotExist(err) {
			findings = append(findings, Finding{
				Code:    "missing-trash-item",
				Message: "active journal record points to a missing trash item",
				Path:    record.TrashedPath,
			})
		}
	}

	return findings, nil
}
