//go:build darwin

package app

import (
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	darwintrash "github.com/Diaszano/bearm/internal/trash/darwin"
)

func newPlatformBackend(home string) domain.TrashBackend {
	return darwintrash.NewBackend(
		darwintrash.NewRootResolver(home, os.Getuid()),
		time.Now,
		id.New,
	)
}
