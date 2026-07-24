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

// Doctor inspects journal and filesystem consistency.
type Doctor struct {
	repository *journal.Repository
}

// NewDoctor creates a diagnostic service.
func NewDoctor(repository *journal.Repository) *Doctor {
	return &Doctor{repository: repository}
}

// Check returns diagnostics without mutating journal or trash.
func (d *Doctor) Check(ctx context.Context) ([]Finding, error) {
	readResult, err := d.repository.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0)
	if readResult.IncompleteTrailingLine {
		findings = append(findings, Finding{
			Code:    "incomplete-journal-line",
			Message: "journal contains an incomplete trailing line",
			Path:    d.repository.Path(),
		})
	}

	active, err := d.repository.ActiveItems(ctx)
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
