//go:build linux

package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func newPlatformBackend(home string) domain.TrashBackend {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if !filepath.IsAbs(dataHome) {
		dataHome = filepath.Join(home, ".local", "share")
	}
	return linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(dataHome, "Trash"),
			UID:       os.Getuid(),
			PerMount:  true,
		},
		time.Now,
		id.New,
	)
}
