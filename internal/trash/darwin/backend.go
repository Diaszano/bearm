//go:build darwin

package darwin

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

// Backend moves macOS targets into system-visible Trash directories.
type Backend struct {
	resolver    RootResolver
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a macOS trash backend.
func NewBackend(resolver RootResolver, clock func() time.Time, idGenerator func() (string, error)) *Backend {
	return &Backend{resolver: resolver, clock: clock, idGenerator: idGenerator}
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "darwin-system-trash"
}

// Resolve selects and prepares the trash root for target.
func (b *Backend) Resolve(ctx context.Context, target domain.PlannedTarget) (domain.Destination, error) {
	if err := ctx.Err(); err != nil {
		return domain.Destination{}, err
	}
	root, err := b.resolver.Resolve(target.AbsolutePath)
	if err != nil {
		return domain.Destination{}, err
	}

	return domain.Destination{
		Root:     root.Path,
		FilesDir: root.Path,
		InfoDir:  filepath.Join(root.Path, ".bearm-info"),
	}, nil
}

// Move atomically moves target into its resolved macOS Trash directory.
func (b *Backend) Move(
	ctx context.Context,
	target domain.PlannedTarget,
	destination domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrashRecord{}, err
	}
	if operationID == "" {
		return domain.TrashRecord{}, errors.New("operation ID is required")
	}

	itemID, err := b.idGenerator()
	if err != nil {
		return domain.TrashRecord{}, err
	}

	deletedAt := b.clock()
	metadata, err := RenderMetadata(Metadata{
		SchemaVersion: 1,
		OriginalPath:  target.AbsolutePath,
		DeletedAt:     deletedAt.UTC(),
	})
	if err != nil {
		return domain.TrashRecord{}, err
	}

	reservation, err := trash.ReserveName(
		destination.FilesDir,
		destination.InfoDir,
		filepath.Base(target.AbsolutePath),
		".json",
		metadata,
	)
	if err != nil {
		return domain.TrashRecord{}, err
	}
	defer func() { _ = reservation.Rollback() }()

	if err := os.Rename(target.AbsolutePath, reservation.TargetPath); err != nil {
		return domain.TrashRecord{}, err
	}

	reservation.Commit()
	return domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        itemID,
		OperationID:   operationID,
		OriginalPath:  target.AbsolutePath,
		TrashedPath:   reservation.TargetPath,
		Backend:       b.Name(),
		DeviceID:      target.DeviceID,
		DeletedAt:     deletedAt.UTC(),
		Status:        "trashed",
	}, nil
}
