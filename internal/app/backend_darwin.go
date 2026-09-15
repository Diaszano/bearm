//go:build darwin

package app

import (
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/pathutil"
	"github.com/Diaszano/bearm/internal/platform"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func newPlatformBackend(home string, dirs platform.Dirs, settings config.Config) domain.TrashBackend {
	_ = dirs
	_ = settings
	return darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		pathutil.NewID,
	)
}
