// Package planner inspects removal operands and builds immutable execution plans.
package planner

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/safety"
)

// IDGenerator creates operation IDs.
type IDGenerator func() (string, error)

// Planner builds removal plans without mutating the filesystem.
type Planner struct {
	policy      *safety.Policy
	idGenerator IDGenerator
}

// New creates a removal planner.
func New(policy *safety.Policy, idGenerator IDGenerator) *Planner {
	return &Planner{policy: policy, idGenerator: idGenerator}
}

// Plan validates and classifies all request operands.
func (p *Planner) Plan(
	ctx context.Context,
	request domain.RemoveRequest,
) (domain.RemovalPlan, []domain.ItemResult) {
	operationID, err := p.idGenerator()
	if err != nil {
		return domain.RemovalPlan{}, []domain.ItemResult{{
			Status: domain.ItemFailed,
			Err:    err,
		}}
	}

	plan := domain.RemovalPlan{
		ID:        operationID,
		CreatedAt: time.Now().UTC(),
		Request:   request,
	}
	failures := make([]domain.ItemResult, 0)
	seen := make(map[string]struct{})
	seenOperands := make(map[string]struct{})

	for _, operand := range request.Operands {
		if err := ctx.Err(); err != nil {
			failures = append(failures, domain.ItemResult{
				Path: operand, Status: domain.ItemFailed, Err: err,
			})
			break
		}

		absolute, err := filepath.Abs(operand)
		if err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		if _, exists := seenOperands[absolute]; exists {
			continue
		}
		seenOperands[absolute] = struct{}{}

		if _, exists := seen[absolute]; exists {
			continue
		}

		if err := p.policy.Check(absolute); err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		info, err := os.Lstat(absolute)
		if err != nil {
			if os.IsNotExist(err) && request.Options.Force {
				continue
			}
			failures = append(failures, failed(operand, err))
			continue
		}

		kind := classify(info)
		if err := ValidatePreserveRootAll(
			absolute,
			kind,
			request.Options.PreserveRoot,
			platform.DeviceID,
		); err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		if kind == domain.TargetDir {
			if !request.Options.Recursive && !request.Options.Directory {
				failures = append(failures, failed(operand, errors.New("is a directory")))
				continue
			}
			if request.Options.Directory && !request.Options.Recursive {
				empty, err := directoryEmpty(absolute)
				if err != nil {
					failures = append(failures, failed(operand, err))
					continue
				}
				if !empty {
					failures = append(failures, failed(operand, errors.New("directory not empty")))
					continue
				}
			}
		}

		deviceID, err := platform.DeviceID(absolute)
		if err != nil {
			failures = append(failures, failed(operand, err))
			continue
		}

		planned := domain.PlannedTarget{
			InputPath:    operand,
			AbsolutePath: absolute,
			Kind:         kind,
			DeviceID:     deviceID,
			RequiresWalk: requiresWalk(kind, request.Options, p.policy),
		}
		if planned.RequiresWalk {
			expanded, skipped, err := ExpandTarget(planned, request.Options, p.policy)
			if err != nil {
				failures = append(failures, failed(operand, err))
				continue
			}
			for _, exp := range expanded {
				if _, exists := seen[exp.AbsolutePath]; exists {
					continue
				}
				seen[exp.AbsolutePath] = struct{}{}
				plan.Targets = append(plan.Targets, exp)
			}
			failures = append(failures, skipped...)
			continue
		}
		seen[absolute] = struct{}{}
		plan.Targets = append(plan.Targets, planned)
	}

	return plan, failures
}

func classify(info os.FileInfo) domain.TargetKind {
	if info.Mode()&os.ModeSymlink != 0 {
		return domain.TargetSymlink
	}
	if info.IsDir() {
		return domain.TargetDir
	}
	if info.Mode().IsRegular() {
		return domain.TargetFile
	}
	return domain.TargetOther
}

func requiresWalk(
	kind domain.TargetKind,
	options domain.RemoveOptions,
	policy *safety.Policy,
) bool {
	if kind != domain.TargetDir {
		return false
	}
	return options.Interactive == domain.InteractiveAlways ||
		options.Verbose ||
		options.OneFileSystem ||
		policy.InspectDescendants()
}

func directoryEmpty(path string) (bool, error) {
	directory, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer directory.Close()

	_, err = directory.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, nil
}

func failed(path string, err error) domain.ItemResult {
	return domain.ItemResult{Path: path, Status: domain.ItemFailed, Err: err}
}
