//go:build linux

package app

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/pathutil"
	"github.com/Diaszano/bearm/internal/platform"
	linuxtrash "github.com/Diaszano/bearm/internal/trash/linux"
)

func newPlatformBackend(home string, dirs platform.Dirs, settings config.Config) domain.TrashBackend {
	_ = home
	return linuxtrash.NewBackend(
		linuxtrash.RootResolver{
			HomeTrash: filepath.Join(filepath.Dir(dirs.DataRoot), "Trash"),
			UID:       os.Getuid(),
			PerMount:  settings.Trash.PerMount,
		},
		time.Now,
		pathutil.NewID,
	)
}
