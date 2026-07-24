// Package journal persists Bearm lifecycle events.
package journal

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// ReadResult contains decoded events and recoverable parse state.
type ReadResult struct {
	Events                 []domain.JournalEvent
	IncompleteTrailingLine bool
}

// Repository stores append-only lifecycle events.
type Repository struct {
	path  string
	clock func() time.Time
}

// New creates a journal repository.
func New(path string, clock func() time.Time) *Repository {
	return &Repository{path: path, clock: clock}
}

// Append implements removal.Journal for completed trash records.
func (r *Repository) Append(ctx context.Context, records []domain.TrashRecord) error {
	events := make([]domain.JournalEvent, 0, len(records))
	now := r.clock()
	for _, record := range records {
		events = append(events, record.Event(domain.JournalTrashed, now))
	}
	return r.AppendEvents(ctx, events)
}

// AppendEvents atomically appends complete JSONL events under a process lock.
func (r *Repository) AppendEvents(ctx context.Context, events []domain.JournalEvent) error {
	if len(events) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(r.path), 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(r.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		_ = file.Close()
	}()

	if err := lockExclusive(file); err != nil {
		return err
	}
	defer func() {
		_ = unlock(file)
	}()

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		if err := encoder.Encode(event); err != nil {
			return err
		}
	}

	if _, err := file.Write(buffer.Bytes()); err != nil {
		return err
	}
	return file.Sync()
}

// ReadAll reads all complete valid events and detects a partial final line.
func (r *Repository) ReadAll(ctx context.Context) (ReadResult, error) {
	file, err := os.Open(r.path)
	if os.IsNotExist(err) {
		return ReadResult{}, nil
	}
	if err != nil {
		return ReadResult{}, err
	}
	defer func() {
		_ = file.Close()
	}()

	if err := lockExclusive(file); err != nil {
		return ReadResult{}, err
	}
	defer func() {
		_ = unlock(file)
	}()

	reader := bufio.NewReader(file)
	result := ReadResult{}
	for {
		if err := ctx.Err(); err != nil {
			return ReadResult{}, err
		}

		line, readErr := reader.ReadBytes('\n')
		if len(line) > 0 {
			if line[len(line)-1] != '\n' {
				result.IncompleteTrailingLine = true
				break
			}

			var event domain.JournalEvent
			if err := json.Unmarshal(bytes.TrimSpace(line), &event); err != nil {
				return ReadResult{}, err
			}
			if err := event.Validate(); err != nil {
				return ReadResult{}, err
			}
			result.Events = append(result.Events, event)
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return ReadResult{}, readErr
		}
	}

	return result, nil
}

// ActiveItems returns the latest active trash record for each item.
func (r *Repository) ActiveItems(ctx context.Context) ([]domain.TrashRecord, error) {
	result, err := r.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	active := make(map[string]domain.TrashRecord)
	for _, event := range result.Events {
		switch event.Action {
		case domain.JournalTrashed:
			active[event.Record.ItemID] = event.Record
		case domain.JournalRestored, domain.JournalPurged:
			delete(active, event.Record.ItemID)
		}
	}

	records := make([]domain.TrashRecord, 0, len(active))
	for _, record := range active {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].DeletedAt.After(records[j].DeletedAt)
	})
	return records, nil
}

// FindItems returns active records matching item IDs.
func (r *Repository) FindItems(ctx context.Context, ids []string) ([]domain.TrashRecord, error) {
	active, err := r.ActiveItems(ctx)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]struct{}, len(ids))
	for _, itemID := range ids {
		wanted[itemID] = struct{}{}
	}

	records := make([]domain.TrashRecord, 0, len(ids))
	for _, record := range active {
		if _, ok := wanted[record.ItemID]; ok {
			records = append(records, record)
		}
	}
	if len(records) != len(wanted) {
		return nil, errors.New("one or more active items were not found")
	}
	return records, nil
}

// LatestOperation returns active records from the latest operation.
func (r *Repository) LatestOperation(ctx context.Context) ([]domain.TrashRecord, error) {
	active, err := r.ActiveItems(ctx)
	if err != nil {
		return nil, err
	}
	if len(active) == 0 {
		return nil, errors.New("no active trash operation found")
	}

	operationID := active[0].OperationID
	records := make([]domain.TrashRecord, 0)
	for _, record := range active {
		if record.OperationID == operationID {
			records = append(records, record)
		}
	}
	return records, nil
}

// Path returns the journal path.
func (r *Repository) Path() string {
	return r.path
}
