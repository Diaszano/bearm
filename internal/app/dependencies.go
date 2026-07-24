package app

import (
	"time"

	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/id"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
	customtrash "github.com/Diaszano/bearm/internal/trash/custom"
)

// Dependencies contains mutable infrastructure used by the application.
type Dependencies struct {
	Backend    domain.TrashBackend
	Journal    removal.Journal
	Repository *journal.Repository
	Policy     *safety.Policy
	Config     config.Config
	ConfigPath string
}

// NewConfiguredBackend creates a platform or custom trash backend.
func NewConfiguredBackend(home string, settings config.Config) (domain.TrashBackend, error) {
	if settings.Trash.CustomPath != "" {
		return customtrash.NewBackend(
			settings.Trash.CustomPath,
			time.Now,
			id.New,
		)
	}
	return newPlatformBackend(home, settings), nil
}
