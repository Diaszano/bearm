// Package restore implements Bearm restore, purge, and doctor use cases.
package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

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
	journal domain.EventAppender
	clock   func() time.Time
}

// NewService creates a restore service.
func NewService(journal domain.EventAppender, clock func() time.Time) *Service {
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

		destination, err := atomicRestore(record.TrashedPath, record.OriginalPath, policy)
		if err != nil {
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

// atomicRestore moves src to dst (or a numbered variant) atomically.
// It uses RenameNoReplace so that the existence check and the move
// happen in a single kernel call, eliminating TOCTOU races.
func atomicRestore(src, dst string, policy CollisionPolicy) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return "", err
	}

	// Fast path: destination does not exist.
	if err := trash.RenameNoReplace(src, dst); err == nil {
		return dst, nil
	} else if !isExist(err) {
		if isCrossDevice(err) {
			return "", fmt.Errorf("restore across filesystems is not supported: source and destination are on different devices")
		}
		return "", err
	}

	// Destination exists — apply collision policy.
	switch policy {
	case CollisionFail:
		return "", errors.New("restore destination already exists")
	case CollisionRename:
		for suffix := 1; suffix < 10000; suffix++ {
			candidate := fmt.Sprintf("%s.restored.%d", dst, suffix)
			if err := trash.RenameNoReplace(src, candidate); err == nil {
				return candidate, nil
			} else if !isExist(err) {
				return "", err
			}
		}
		return "", errors.New("restore destination limit exceeded")
	case CollisionOverwrite:
		if err := os.RemoveAll(dst); err != nil {
			return "", err
		}
		if err := os.Rename(src, dst); err != nil {
			return "", err
		}
		return dst, nil
	default:
		return "", errors.New("unsupported restore collision policy")
	}
}

// isExist reports whether err indicates the target already exists.
func isExist(err error) bool {
	return errors.Is(err, syscall.EEXIST) || os.IsExist(err)
}

func failed(path string, err error) domain.ItemResult {
	return domain.ItemResult{Path: path, Status: domain.ItemFailed, Err: err}
}

func isCrossDevice(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}
