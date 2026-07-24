// Package restore implements Bearm restore, purge, and doctor use cases.
package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// EventAppender appends lifecycle events.
type EventAppender interface {
	AppendEvents(context.Context, []domain.JournalEvent) error
}

// CollisionPolicy controls restore destination collisions.
type CollisionPolicy string

const (
	// CollisionFail rejects an existing destination.
	CollisionFail CollisionPolicy = "fail"
	// CollisionRename restores to a unique sibling path.
	CollisionRename CollisionPolicy = "rename"
	// CollisionOverwrite permanently removes the destination before restore.
	CollisionOverwrite CollisionPolicy = "overwrite"
)

// Service restores active trash records.
type Service struct {
	journal EventAppender
	clock   func() time.Time
}

// NewService creates a restore service.
func NewService(journal EventAppender, clock func() time.Time) *Service {
	return &Service{journal: journal, clock: clock}
}

// Restore restores records independently and appends lifecycle events.
func (s *Service) Restore(
	ctx context.Context,
	records []domain.TrashRecord,
	policy CollisionPolicy,
) []domain.ItemResult {
	results := make([]domain.ItemResult, 0, len(records))
	events := make([]domain.JournalEvent, 0, len(records))

	for _, record := range records {
		if err := ctx.Err(); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			break
		}

		destination, err := resolveDestination(record.OriginalPath, policy)
		if err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}
		if err := os.Rename(record.TrashedPath, destination); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}

		restored := record
		restored.OriginalPath = destination
		restored.Status = "restored"
		events = append(events, restored.Event(domain.JournalRestored, s.clock()))
		results = append(results, domain.ItemResult{
			Path:   destination,
			Status: domain.ItemRestored,
			Record: &restored,
		})
	}

	if len(events) > 0 {
		if err := s.journal.AppendEvents(ctx, events); err != nil {
			results = append(results, failed("", fmt.Errorf("append restore journal: %w", err)))
		}
	}

	return results
}

func resolveDestination(path string, policy CollisionPolicy) (string, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return path, nil
	}
	if err != nil {
		return "", err
	}

	switch policy {
	case CollisionFail:
		return "", errors.New("restore destination already exists")
	case CollisionRename:
		for suffix := 1; suffix < 10000; suffix++ {
			candidate := fmt.Sprintf("%s.restored.%d", path, suffix)
			if _, err := os.Lstat(candidate); os.IsNotExist(err) {
				return candidate, nil
			}
		}
		return "", errors.New("restore destination limit exceeded")
	case CollisionOverwrite:
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", errors.New("unsupported restore collision policy")
	}
}

func failed(path string, err error) domain.ItemResult {
	return domain.ItemResult{Path: path, Status: domain.ItemFailed, Err: err}
}
