package testutil

import (
	"context"

	"github.com/Diaszano/bearm/internal/domain"
)

// Journal is an in-memory journal test double.
type Journal struct {
	Records []domain.TrashRecord
	Err     error
}

// Append stores operation records.
func (j *Journal) Append(_ context.Context, records []domain.TrashRecord) error {
	j.Records = append(j.Records, records...)
	return j.Err
}
