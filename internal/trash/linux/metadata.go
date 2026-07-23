//go:build linux

package linux

import (
	"fmt"
	"time"

	"github.com/Diaszano/bearm/internal/trash"
)

// RenderTrashInfo renders a FreeDesktop .trashinfo document.
func RenderTrashInfo(path string, deletedAt time.Time) []byte {
	return []byte(fmt.Sprintf(
		"[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		trash.EncodePath(path),
		deletedAt.Format("2006-01-02T15:04:05"),
	))
}
