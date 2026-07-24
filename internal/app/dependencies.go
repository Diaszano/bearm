package app

import (
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/safety"
)

// Dependencies contains mutable infrastructure used by the application.
type Dependencies struct {
	Backend domain.TrashBackend
	Journal removal.Journal
	Policy  *safety.Policy
}
