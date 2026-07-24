// Package custom implements a same-filesystem user-selected trash.
package custom

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/trash"
)

// Backend moves targets into a configured trash root.
type Backend struct {
	root        string
	clock       func() time.Time
	idGenerator func() (string, error)
}

// NewBackend creates a custom trash backend.
func NewBackend(root string, clock func() time.Time, idGenerator func() (string, error)) (*Backend, error) {
	if !filepath.IsAbs(root) {
		return nil, errors.New("custom trash root must be absolute")
	}
	if clock == nil {
		return nil, errors.New("clock function must be specified")
	}
	if idGenerator == nil {
		return nil, errors.New("id generator function must be specified")
	}
	return &Backend{root: filepath.Clean(root), clock: clock, idGenerator: idGenerator}, nil
}

// Name returns the backend identifier.
func (b *Backend) Name() string {
	return "custom-trash"
}

// Resolve creates and validates the custom trash skeleton.
func (b *Backend) Resolve(_ context.Context, _ domain.PlannedTarget) (domain.Destination, error) {
	filesDir := filepath.Join(b.root, "files")
	infoDir := filepath.Join(b.root, "info")
	for _, directory := range []string{b.root, filesDir, infoDir} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			return domain.Destination{}, err
		}
		info, err := os.Lstat(directory)
		if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return domain.Destination{}, errors.New("custom trash contains an unsafe directory")
		}
	}
	return domain.Destination{
		Root: b.root, FilesDir: filesDir, InfoDir: infoDir,
	}, nil
}

// Move renames one target into the custom trash.
func (b *Backend) Move(
	ctx context.Context,
	target domain.PlannedTarget,
	destination domain.Destination,
	operationID string,
) (domain.TrashRecord, error) {
	if err := ctx.Err(); err != nil {
		return domain.TrashRecord{}, err
	}

	itemID, err := b.idGenerator()
	if err != nil {
		return domain.TrashRecord{}, err
	}

	deletedAt := b.clock()
	metadata, err := json.Marshal(struct {
		SchemaVersion int       `json:"schema_version"`
		OriginalPath  string    `json:"original_path"`
		DeletedAt     time.Time `json:"deleted_at"`
	}{
		SchemaVersion: 1,
		OriginalPath:  target.AbsolutePath,
		DeletedAt:     deletedAt.UTC(),
	})
	if err != nil {
		return domain.TrashRecord{}, err
	}
	metadata = append(metadata, '\n')

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
		ItemID:        itemID, OperationID: operationID,
		OriginalPath: target.AbsolutePath, TrashedPath: reservation.TargetPath,
		Backend: b.Name(), DeviceID: target.DeviceID,
		DeletedAt: deletedAt.UTC(), Status: "trashed",
	}, nil
}
