package app

import (
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
)

// Dependencies contains mutable infrastructure used by the application.
type Dependencies struct {
	Backend    domain.TrashBackend
	Journal    removal.Journal
	Repository *journal.Repository
	Policy     *safety.Policy
}

// NewPlatformBackend creates the host trash backend.
func NewPlatformBackend(home string) domain.TrashBackend {
	return newPlatformBackend(home)
}
