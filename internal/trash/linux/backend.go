//go:build linux

// Package linux implements the FreeDesktop Trash backend.
package linux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

// Backend moves Linux targets into FreeDesktop trash.
type Backend struct {
	resolver    RootResolver
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a Linux trash backend.
func NewBackend(resolver RootResolver, clock func() time.Time, idGenerator func() (string, error)) *Backend {
	return &Backend{
		resolver:    resolver,
		clock:       clock,
		idGenerator: idGenerator,
	}
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "linux-freedesktop"
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
		FilesDir: filepath.Join(root.Path, "files"),
		InfoDir:  filepath.Join(root.Path, "info"),
	}, nil
}

// Move atomically moves target into its resolved trash root.
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

	root := Root{
		Path:          destination.Root,
		UID:           b.resolver.UID,
		CheckSecurity: destination.Root != b.resolver.HomeTrash,
	}
	if err := root.Ensure(); err != nil {
		return domain.TrashRecord{}, err
	}

	deletedAt := b.clock()
	infoPathValue := target.AbsolutePath
	resolvedRoot, err := b.resolver.Resolve(target.AbsolutePath)
	if err != nil {
		return domain.TrashRecord{}, err
	}
	if resolvedRoot.RelativeInfoPath {
		relative, relErr := filepath.Rel(resolvedRoot.MountPoint, target.AbsolutePath)
		if relErr != nil {
			return domain.TrashRecord{}, relErr
		}
		infoPathValue = relative
	}

	reservation, err := trash.ReserveName(
		filepath.Join(root.Path, "files"),
		filepath.Join(root.Path, "info"),
		filepath.Base(target.AbsolutePath),
		".trashinfo",
		RenderTrashInfo(infoPathValue, deletedAt.Local()),
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
