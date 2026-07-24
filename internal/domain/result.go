package domain

// ItemStatus identifies one compatibility removal outcome.
type ItemStatus string

const (
	// ItemTrashed indicates a successful move to trash.
	ItemTrashed ItemStatus = "trashed"
	// ItemRestored indicates a successful restore.
	ItemRestored ItemStatus = "restored"
	// ItemPurged indicates a successful permanent purge.
	ItemPurged ItemStatus = "purged"
	// ItemSkipped indicates an intentionally ignored target.
	ItemSkipped ItemStatus = "skipped"
	// ItemDeclined indicates a user-declined interactive target.
	ItemDeclined ItemStatus = "declined"
	// ItemFailed indicates an operational failure.
	ItemFailed ItemStatus = "failed"
)

// ItemResult contains one target outcome.
type ItemResult struct {
	Path   string
	Status ItemStatus
	Record *TrashRecord
	Err    error
}

// RemovalResult contains the outcomes of one compatibility operation.
type RemovalResult struct {
	OperationID string
	Items       []ItemResult
}

// HasFailures reports whether any target failed.
func (r RemovalResult) HasFailures() bool {
	for _, item := range r.Items {
		if item.Status == ItemFailed {
			return true
		}
	}
	return false
}

// ExitCode returns the compatibility operational exit code.
func (r RemovalResult) ExitCode(_ CompatibilityProfile) int {
	if r.HasFailures() {
		return 1
	}
	return 0
}
