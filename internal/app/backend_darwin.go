//go:build darwin

package app

import (
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func newPlatformBackend(home string, settings config.Config) domain.TrashBackend {
	_ = settings
	return darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		id.New,
	)
}
