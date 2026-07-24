package removal

import (
	"context"
	"fmt"
	"io"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/safety"
)

// Journal appends durable trash records.
type Journal interface {
	Append(context.Context, []domain.TrashRecord) error
}

// Executor moves planned targets and records completed operations.
type Executor struct {
	backend  domain.TrashBackend
	journal  Journal
	policy   *safety.Policy
	prompter *Prompter
	verbose  io.Writer
}

// NewExecutor creates a removal executor.
func NewExecutor(
	backend domain.TrashBackend,
	journal Journal,
	policy *safety.Policy,
	prompter *Prompter,
	verbose io.Writer,
) *Executor {
	if verbose == nil {
		verbose = io.Discard
	}
	return &Executor{
		backend: backend, journal: journal, policy: policy, prompter: prompter, verbose: verbose,
	}
}

// Execute performs one immutable removal plan.
func (e *Executor) Execute(ctx context.Context, plan domain.RemovalPlan) domain.RemovalResult {
	result := domain.RemovalResult{OperationID: plan.ID}
	if err := ctx.Err(); err != nil {
		result.Items = append(result.Items, domain.ItemResult{Status: domain.ItemFailed, Err: err})
		return result
	}

	if NeedsOncePrompt(plan) {
		accepted, err := e.prompter.ConfirmOnce(plan)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{Status: domain.ItemFailed, Err: err})
			return result
		}
		if !accepted {
			for _, target := range plan.Targets {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemDeclined,
				})
			}
			return result
		}
	}

	records := make([]domain.TrashRecord, 0, len(plan.Targets))
	for _, target := range plan.Targets {
		if err := ctx.Err(); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			break
		}
		if err := e.policy.Check(target.AbsolutePath); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}

		if plan.Request.Options.Interactive == domain.InteractiveAlways {
			accepted, err := e.prompter.ConfirmTarget(target.InputPath)
			if err != nil {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemFailed, Err: err,
				})
				continue
			}
			if !accepted {
				result.Items = append(result.Items, domain.ItemResult{
					Path: target.InputPath, Status: domain.ItemDeclined,
				})
				continue
			}
		}

		destination, err := e.backend.Resolve(ctx, target)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}
		record, err := e.backend.Move(ctx, target, destination, plan.ID)
		if err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Path: target.InputPath, Status: domain.ItemFailed, Err: err,
			})
			continue
		}

		records = append(records, record)
		recordCopy := record
		result.Items = append(result.Items, domain.ItemResult{
			Path: target.InputPath, Status: domain.ItemTrashed, Record: &recordCopy,
		})
		if plan.Request.Options.Verbose {
			fmt.Fprintln(e.verbose, target.InputPath)
		}
	}

	if len(records) > 0 {
		if err := e.journal.Append(ctx, records); err != nil {
			result.Items = append(result.Items, domain.ItemResult{
				Status: domain.ItemFailed,
				Err:    fmt.Errorf("append operation journal: %w", err),
			})
		}
	}

	return result
}
