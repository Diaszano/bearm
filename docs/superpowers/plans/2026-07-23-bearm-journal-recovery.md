# Bearm Journal and Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a durable append-only journal and native commands for listing, restoring, permanently purging, and diagnosing Bearm trash records.

**Architecture:** The journal stores versioned JSONL events under an exclusive advisory lock. Restore and purge are explicit use cases that append new events instead of mutating history. Native command rendering remains separate from storage and filesystem behavior.

**Tech Stack:** Go 1.24+, Go standard library, `golang.org/x/sys/unix`, JSONL, table-driven and concurrency tests.

## Global Constraints

- Module path is exactly `github.com/Diaszano/bearm`.
- Source code, identifiers, comments, technical documentation, branches, and commits are in English.
- Native user-facing output defaults to Brazilian Portuguese.
- Journal storage is append-only and schema-versioned.
- One complete JSON object is written per line.
- Journal appends use an exclusive process lock and `fsync`.
- Incomplete trailing lines are ignored for normal reads and reported by `doctor`.
- Restore defaults to collision failure.
- Permanent deletion is available only through native purge with confirmation or `--yes`.
- Restored and purged records append new events; historical records are never rewritten.
- No file contents are stored in the journal.
- Use strict Go typing; do not use `any` in production APIs.
- Public Go declarations require standard Go documentation comments.
- Use TDD and one focused Conventional Commit per task.
- Do not add AI attribution trailers.
- Go version floor is `1.24.0`.

---

## File Structure

```text
internal/
├── app/
│   ├── app.go
│   ├── app_test.go
│   └── dependencies.go
├── domain/
│   ├── journal.go
│   └── journal_test.go
├── journal/
│   ├── repository.go
│   ├── repository_test.go
│   ├── lock_unix.go
│   └── recovery_test.go
├── restore/
│   ├── service.go
│   ├── service_test.go
│   ├── purge.go
│   ├── purge_test.go
│   ├── doctor.go
│   └── doctor_test.go
└── testutil/
    └── journal.go
```

## Dependency Order

```text
Task 1 journal domain events
  └── Task 2 locked JSONL repository
          ├── Task 3 restore service
          ├── Task 4 purge service
          ├── Task 5 doctor service
          └── Task 6 native command integration
```

### Task 1: Define Versioned Journal Events

**Files:**
- Create: `internal/domain/journal.go`
- Create: `internal/domain/journal_test.go`
- Modify: `internal/domain/trash.go`

**Interfaces:**
- Consumes: `domain.TrashRecord`.
- Produces:
  - `domain.JournalAction`
  - `domain.JournalEvent`
  - `domain.TrashRecord.Event(action, occurredAt)`
  - `domain.JournalEvent.Validate() error`

- [ ] **Step 1: Write failing event tests**

Create `internal/domain/journal_test.go`:

```go
package domain_test

import (
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestTrashRecordEventCreatesVersionedEvent(t *testing.T) {
	t.Parallel()

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file.txt",
		TrashedPath:   "/trash/file.txt",
		Backend:       "fake",
		DeletedAt:     time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
		Status:        "trashed",
	}
	occurredAt := time.Date(2026, 7, 23, 12, 0, 1, 0, time.UTC)

	event := record.Event(domain.JournalTrashed, occurredAt)
	if event.SchemaVersion != 1 || event.Action != domain.JournalTrashed {
		t.Fatalf("event = %#v", event)
	}
	if event.Record.ItemID != "item-1" || !event.OccurredAt.Equal(occurredAt) {
		t.Fatalf("event = %#v", event)
	}
}

func TestJournalEventValidateRejectsUnknownAction(t *testing.T) {
	t.Parallel()

	event := domain.JournalEvent{
		SchemaVersion: 1,
		Action:        domain.JournalAction("unknown"),
		OccurredAt:    time.Now(),
		Record:        domain.TrashRecord{ItemID: "item-1", OperationID: "operation-1"},
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestJournalEventValidateRejectsMissingIdentity(t *testing.T) {
	t.Parallel()

	event := domain.JournalEvent{
		SchemaVersion: 1,
		Action:        domain.JournalTrashed,
		OccurredAt:    time.Now(),
	}
	if err := event.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
```

- [ ] **Step 2: Implement journal events**

Create `internal/domain/journal.go`:

```go
package domain

import (
	"errors"
	"time"
)

// JournalAction identifies a durable lifecycle event.
type JournalAction string

const (
	// JournalTrashed records a completed move to trash.
	JournalTrashed JournalAction = "trashed"
	// JournalRestored records a completed restore.
	JournalRestored JournalAction = "restored"
	// JournalPurged records a completed permanent purge.
	JournalPurged JournalAction = "purged"
)

// JournalEvent is one append-only lifecycle event.
type JournalEvent struct {
	SchemaVersion int           `json:"schema_version"`
	Action        JournalAction `json:"action"`
	OccurredAt    time.Time     `json:"occurred_at"`
	Record        TrashRecord   `json:"record"`
}

// Validate validates journal invariants.
func (e JournalEvent) Validate() error {
	if e.SchemaVersion != 1 {
		return errors.New("unsupported journal schema version")
	}
	switch e.Action {
	case JournalTrashed, JournalRestored, JournalPurged:
	default:
		return errors.New("unsupported journal action")
	}
	if e.Record.ItemID == "" || e.Record.OperationID == "" {
		return errors.New("journal event requires item and operation IDs")
	}
	if e.OccurredAt.IsZero() {
		return errors.New("journal event requires occurrence time")
	}
	return nil
}
```

Add to `internal/domain/trash.go`:

```go
// Event converts a trash record into a versioned journal event.
func (r TrashRecord) Event(action JournalAction, occurredAt time.Time) JournalEvent {
	return JournalEvent{
		SchemaVersion: 1,
		Action:        action,
		OccurredAt:    occurredAt.UTC(),
		Record:        r,
	}
}
```

- [ ] **Step 3: Run domain tests**

Run:

```bash
gofmt -w internal/domain
go test ./internal/domain -v
```

Expected:

```text
PASS
```

- [ ] **Step 4: Commit**

```bash
git add internal/domain
git commit -m "feat: define Bearm journal events"
```

### Task 2: Implement a Locked Append-Only JSONL Repository

**Files:**
- Create: `internal/journal/lock_unix.go`
- Create: `internal/journal/repository.go`
- Create: `internal/journal/repository_test.go`
- Create: `internal/journal/recovery_test.go`
- Modify: `internal/testutil/journal.go`

**Interfaces:**
- Consumes: state root, `domain.JournalEvent`, `domain.TrashRecord`.
- Produces:
  - `journal.Repository`
  - `journal.New(path, clock)`
  - `(*Repository).Append(ctx, records) error`
  - `(*Repository).AppendEvents(ctx, events) error`
  - `(*Repository).ReadAll(ctx) (ReadResult, error)`
  - `(*Repository).ActiveItems(ctx) ([]domain.TrashRecord, error)`
  - `(*Repository).FindItems(ctx, ids) ([]domain.TrashRecord, error)`
  - `(*Repository).LatestOperation(ctx) ([]domain.TrashRecord, error)`

- [ ] **Step 1: Write failing append/read tests**

Create `internal/journal/repository_test.go`:

```go
package journal_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
)

func TestAppendAndReadAll(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	clock := func() time.Time {
		return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	}
	repository := journal.New(path, clock)

	records := []domain.TrashRecord{
		{
			SchemaVersion: 1,
			ItemID:        "item-1",
			OperationID:   "operation-1",
			OriginalPath:  "/work/a",
			TrashedPath:   "/trash/a",
			Backend:       "fake",
			DeletedAt:     clock(),
			Status:        "trashed",
		},
		{
			SchemaVersion: 1,
			ItemID:        "item-2",
			OperationID:   "operation-1",
			OriginalPath:  "/work/b",
			TrashedPath:   "/trash/b",
			Backend:       "fake",
			DeletedAt:     clock(),
			Status:        "trashed",
		},
	}

	if err := repository.Append(context.Background(), records); err != nil {
		t.Fatalf("Append() error = %v", err)
	}

	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(result.Events) != 2 || result.IncompleteTrailingLine {
		t.Fatalf("result = %#v", result)
	}
}

func TestActiveItemsAppliesLifecycleEvents(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	clock := func() time.Time {
		return time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	}
	repository := journal.New(path, clock)

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/a",
		TrashedPath:   "/trash/a",
		Backend:       "fake",
		DeletedAt:     clock(),
		Status:        "trashed",
	}
	events := []domain.JournalEvent{
		record.Event(domain.JournalTrashed, clock()),
		record.Event(domain.JournalRestored, clock().Add(time.Second)),
	}

	if err := repository.AppendEvents(context.Background(), events); err != nil {
		t.Fatal(err)
	}

	active, err := repository.ActiveItems(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Fatalf("active = %#v, want empty", active)
	}
}
```

- [ ] **Step 2: Write failing recovery tests**

Create `internal/journal/recovery_test.go`:

```go
package journal_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/journal"
)

func TestReadAllIgnoresIncompleteTrailingLine(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	content := `{"schema_version":1,"action":"trashed","occurred_at":"2026-07-23T12:00:00Z","record":{"schema_version":1,"item_id":"item-1","operation_id":"operation-1","original_path":"/work/a","trashed_path":"/trash/a","backend":"fake","device_id":1,"deleted_at":"2026-07-23T12:00:00Z","status":"trashed"}}` + "\n" +
		`{"schema_version":1,"action":"trashed"`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := journal.New(path, time.Now)
	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if len(result.Events) != 1 || !result.IncompleteTrailingLine {
		t.Fatalf("result = %#v", result)
	}
}
```

- [ ] **Step 3: Implement Unix file locking**

Create `internal/journal/lock_unix.go`:

```go
//go:build linux || darwin

package journal

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockExclusive(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX)
}

func unlock(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
```

- [ ] **Step 4: Implement the JSONL repository**

Create `internal/journal/repository.go`:

```go
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
	defer file.Close()

	if err := lockExclusive(file); err != nil {
		return err
	}
	defer unlock(file)

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
	defer file.Close()

	if err := lockExclusive(file); err != nil {
		return ReadResult{}, err
	}
	defer unlock(file)

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
```

- [ ] **Step 5: Run journal tests**

Run:

```bash
gofmt -w internal/journal
go test ./internal/journal -v
go test -race ./internal/journal -count=10
```

Expected:

```text
PASS
```

- [ ] **Step 6: Add concurrent append coverage**

Append to `internal/journal/repository_test.go`:

```go
func TestConcurrentAppendPreservesCompleteLines(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)

	const count = 25
	errs := make(chan error, count)
	for index := 0; index < count; index++ {
		index := index
		go func() {
			record := domain.TrashRecord{
				SchemaVersion: 1,
				ItemID:        fmt.Sprintf("item-%d", index),
				OperationID:   fmt.Sprintf("operation-%d", index),
				OriginalPath:  fmt.Sprintf("/work/%d", index),
				TrashedPath:   fmt.Sprintf("/trash/%d", index),
				Backend:       "fake",
				DeletedAt:     time.Now(),
				Status:        "trashed",
			}
			errs <- repository.Append(context.Background(), []domain.TrashRecord{record})
		}()
	}

	for index := 0; index < count; index++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}

	result, err := repository.ReadAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Events) != count {
		t.Fatalf("events = %d, want %d", len(result.Events), count)
	}
}
```

Add import:

```go
"fmt"
```

Run:

```bash
go test -race ./internal/journal -run TestConcurrentAppendPreservesCompleteLines -count=10
```

Expected:

```text
PASS with no race reports.
```

- [ ] **Step 7: Update the in-memory test journal**

Replace `internal/testutil/journal.go` with:

```go
package testutil

import (
	"context"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Journal is an in-memory journal test double.
type Journal struct {
	Records []domain.TrashRecord
	Events  []domain.JournalEvent
	Err     error
}

// Append stores trash events.
func (j *Journal) Append(_ context.Context, records []domain.TrashRecord) error {
	j.Records = append(j.Records, records...)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	for _, record := range records {
		j.Events = append(j.Events, record.Event(domain.JournalTrashed, now))
	}
	return j.Err
}

// AppendEvents stores lifecycle events.
func (j *Journal) AppendEvents(_ context.Context, events []domain.JournalEvent) error {
	j.Events = append(j.Events, events...)
	return j.Err
}
```

- [ ] **Step 8: Commit**

```bash
git add internal/journal internal/testutil/journal.go
git commit -m "feat: persist Bearm operation journal"
```

### Task 3: Restore Active Trash Items

**Files:**
- Create: `internal/restore/service.go`
- Create: `internal/restore/service_test.go`

**Interfaces:**
- Consumes:
  - active `domain.TrashRecord` values;
  - append-events journal;
  - collision policy.
- Produces:
  - `restore.CollisionPolicy`
  - `restore.Service`
  - `(*Service).Restore(ctx, records, policy) []domain.ItemResult`

- [ ] **Step 1: Write failing restore tests**

Create `internal/restore/service_test.go`:

```go
package restore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/restore"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestRestoreMovesItemBackAndAppendsEvent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	if err := os.MkdirAll(filepath.Dir(trashed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	journal := &testutil.Journal{}
	service := restore.NewService(journal, func() time.Time {
		return time.Date(2026, 7, 23, 13, 0, 0, 0, time.UTC)
	})

	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}

	results := service.Restore(context.Background(), []domain.TrashRecord{record}, restore.CollisionFail)
	if len(results) != 1 || results[0].Status != domain.ItemTrashed {
		t.Fatalf("results = %#v", results)
	}
	if got, err := os.ReadFile(original); err != nil || string(got) != "data" {
		t.Fatalf("original data = %q, error = %v", got, err)
	}
	if len(journal.Events) != 1 || journal.Events[0].Action != domain.JournalRestored {
		t.Fatalf("events = %#v", journal.Events)
	}
}

func TestRestoreFailsOnDestinationCollision(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	for _, path := range []string{trashed, original} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	service := restore.NewService(&testutil.Journal{}, time.Now)
	results := service.Restore(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
	}}, restore.CollisionFail)

	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("results = %#v", results)
	}
}
```

- [ ] **Step 2: Implement restore service**

Create `internal/restore/service.go`:

```go
// Package restore implements Bearm restore, purge, and doctor use cases.
package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// EventAppender appends lifecycle events.
type EventAppender interface {
	AppendEvents(context.Context, []domain.JournalEvent) error
}

// CollisionPolicy controls restore destination collisions.
type CollisionPolicy string

const (
	// CollisionFail rejects an existing destination.
	CollisionFail CollisionPolicy = "fail"
	// CollisionRename restores to a unique sibling path.
	CollisionRename CollisionPolicy = "rename"
	// CollisionOverwrite permanently removes the destination before restore.
	CollisionOverwrite CollisionPolicy = "overwrite"
)

// Service restores active trash records.
type Service struct {
	journal EventAppender
	clock   func() time.Time
}

// NewService creates a restore service.
func NewService(journal EventAppender, clock func() time.Time) *Service {
	return &Service{journal: journal, clock: clock}
}

// Restore restores records independently and appends lifecycle events.
func (s *Service) Restore(
	ctx context.Context,
	records []domain.TrashRecord,
	policy CollisionPolicy,
) []domain.ItemResult {
	results := make([]domain.ItemResult, 0, len(records))
	events := make([]domain.JournalEvent, 0, len(records))

	for _, record := range records {
		if err := ctx.Err(); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			break
		}

		destination, err := resolveDestination(record.OriginalPath, policy)
		if err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}
		if err := os.Rename(record.TrashedPath, destination); err != nil {
			results = append(results, failed(record.OriginalPath, err))
			continue
		}

		restored := record
		restored.OriginalPath = destination
		restored.Status = "restored"
		events = append(events, restored.Event(domain.JournalRestored, s.clock()))
		results = append(results, domain.ItemResult{
			Path: destination,
			Status: domain.ItemTrashed,
			Record: &restored,
		})
	}

	if len(events) > 0 {
		if err := s.journal.AppendEvents(ctx, events); err != nil {
			results = append(results, failed("", fmt.Errorf("append restore journal: %w", err)))
		}
	}

	return results
}

func resolveDestination(path string, policy CollisionPolicy) (string, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return path, nil
	}
	if err != nil {
		return "", err
	}

	switch policy {
	case CollisionFail:
		return "", errors.New("restore destination already exists")
	case CollisionRename:
		for suffix := 1; suffix < 10000; suffix++ {
			candidate := fmt.Sprintf("%s.restored.%d", path, suffix)
			if _, err := os.Lstat(candidate); os.IsNotExist(err) {
				return candidate, nil
			}
		}
		return "", errors.New("restore destination limit exceeded")
	case CollisionOverwrite:
		if err := os.RemoveAll(path); err != nil {
			return "", err
		}
		return path, nil
	default:
		return "", errors.New("unsupported restore collision policy")
	}
}

func failed(path string, err error) domain.ItemResult {
	return domain.ItemResult{Path: path, Status: domain.ItemFailed, Err: err}
}
```

- [ ] **Step 3: Correct the restored result status**

Add a new status to `internal/domain/result.go`:

```go
// ItemRestored indicates a successful restore.
ItemRestored ItemStatus = "restored"
```

Replace in `internal/restore/service.go`:

```go
Status: domain.ItemTrashed,
```

with:

```go
Status: domain.ItemRestored,
```

Update the success assertion in `internal/restore/service_test.go` to expect `domain.ItemRestored`.

- [ ] **Step 4: Add rename collision coverage**

Append to `internal/restore/service_test.go`:

```go
func TestRestoreRenameCreatesUniqueDestination(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file.txt")
	original := filepath.Join(root, "work", "file.txt")
	for _, path := range []string{trashed, original} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(path), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	service := restore.NewService(&testutil.Journal{}, time.Now)
	results := service.Restore(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  original,
		TrashedPath:   trashed,
	}}, restore.CollisionRename)

	if len(results) != 1 || results[0].Status != domain.ItemRestored {
		t.Fatalf("results = %#v", results)
	}
	if results[0].Path != original+".restored.1" {
		t.Fatalf("restored path = %q", results[0].Path)
	}
}
```

- [ ] **Step 5: Run restore tests**

Run:

```bash
gofmt -w internal/domain internal/restore
go test ./internal/restore -run TestRestore -v
```

Expected:

```text
PASS
```

- [ ] **Step 6: Commit**

```bash
git add internal/domain/result.go internal/restore/service.go internal/restore/service_test.go
git commit -m "feat: restore Bearm trash items"
```

### Task 4: Add Explicit Permanent Purge

**Files:**
- Create: `internal/restore/purge.go`
- Create: `internal/restore/purge_test.go`
- Modify: `internal/domain/result.go`

**Interfaces:**
- Consumes: selected active records and confirmed native command context.
- Produces:
  - `restore.Purger`
  - `(*Purger).Purge(ctx, records, confirmed) []domain.ItemResult`

- [ ] **Step 1: Add purge status**

Add to `internal/domain/result.go`:

```go
// ItemPurged indicates a successful permanent purge.
ItemPurged ItemStatus = "purged"
```

- [ ] **Step 2: Write failing purge tests**

Create `internal/restore/purge_test.go`:

```go
package restore_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/restore"
	"github.com/Diaszano/bearm/internal/testutil"
)

func TestPurgeRequiresConfirmation(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	purger := restore.NewPurger(&testutil.Journal{}, time.Now)
	results := purger.Purge(context.Background(), []domain.TrashRecord{{
		ItemID: "item-1", OperationID: "operation-1", TrashedPath: path,
	}}, false)

	if len(results) != 1 || results[0].Status != domain.ItemFailed {
		t.Fatalf("results = %#v", results)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should remain: %v", err)
	}
}

func TestPurgeRemovesTargetAndAppendsEvent(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "directory")
	if err := os.MkdirAll(filepath.Join(path, "sub"), 0o700); err != nil {
		t.Fatal(err)
	}

	journal := &testutil.Journal{}
	purger := restore.NewPurger(journal, time.Now)
	results := purger.Purge(context.Background(), []domain.TrashRecord{{
		SchemaVersion: 1,
		ItemID: "item-1", OperationID: "operation-1", TrashedPath: path,
	}}, true)

	if len(results) != 1 || results[0].Status != domain.ItemPurged {
		t.Fatalf("results = %#v", results)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("path still exists: %v", err)
	}
	if len(journal.Events) != 1 || journal.Events[0].Action != domain.JournalPurged {
		t.Fatalf("events = %#v", journal.Events)
	}
}
```

- [ ] **Step 3: Implement the purger**

Create `internal/restore/purge.go`:

```go
package restore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
)

// Purger permanently removes explicitly selected trash items.
type Purger struct {
	journal EventAppender
	clock   func() time.Time
}

// NewPurger creates a permanent purge service.
func NewPurger(journal EventAppender, clock func() time.Time) *Purger {
	return &Purger{journal: journal, clock: clock}
}

// Purge permanently removes records only when confirmed is true.
func (p *Purger) Purge(
	ctx context.Context,
	records []domain.TrashRecord,
	confirmed bool,
) []domain.ItemResult {
	if !confirmed {
		results := make([]domain.ItemResult, 0, len(records))
		for _, record := range records {
			results = append(results, failed(record.TrashedPath, errors.New("purge confirmation required")))
		}
		return results
	}

	results := make([]domain.ItemResult, 0, len(records))
	events := make([]domain.JournalEvent, 0, len(records))
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			results = append(results, failed(record.TrashedPath, err))
			break
		}
		if err := os.RemoveAll(record.TrashedPath); err != nil {
			results = append(results, failed(record.TrashedPath, err))
			continue
		}

		purged := record
		purged.Status = "purged"
		events = append(events, purged.Event(domain.JournalPurged, p.clock()))
		results = append(results, domain.ItemResult{
			Path: record.TrashedPath, Status: domain.ItemPurged, Record: &purged,
		})
	}

	if len(events) > 0 {
		if err := p.journal.AppendEvents(ctx, events); err != nil {
			results = append(results, failed("", fmt.Errorf("append purge journal: %w", err)))
		}
	}

	return results
}
```

- [ ] **Step 4: Run purge tests**

Run:

```bash
gofmt -w internal/domain internal/restore
go test ./internal/restore -run TestPurge -v
```

Expected:

```text
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/domain/result.go internal/restore/purge.go internal/restore/purge_test.go
git commit -m "feat: purge Bearm trash items explicitly"
```

### Task 5: Diagnose Journal and Filesystem Inconsistencies

**Files:**
- Create: `internal/restore/doctor.go`
- Create: `internal/restore/doctor_test.go`

**Interfaces:**
- Consumes: `journal.Repository`.
- Produces:
  - `restore.Doctor`
  - `restore.Finding`
  - `(*Doctor).Check(ctx) ([]Finding, error)`

- [ ] **Step 1: Write failing doctor tests**

Create `internal/restore/doctor_test.go`:

```go
package restore_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/restore"
)

func TestDoctorReportsMissingTrashedPath(t *testing.T) {
	t.Parallel()

	repository := journal.New(filepath.Join(t.TempDir(), "journal.jsonl"), time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file",
		TrashedPath:   filepath.Join(t.TempDir(), "missing"),
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	doctor := restore.NewDoctor(repository)
	findings, err := doctor.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Code != "missing-trash-item" {
		t.Fatalf("findings = %#v", findings)
	}
}

func TestDoctorReportsIncompleteTrailingLine(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	if err := os.WriteFile(path, []byte(`{"schema_version":1`), 0o600); err != nil {
		t.Fatal(err)
	}

	doctor := restore.NewDoctor(journal.New(path, time.Now))
	findings, err := doctor.Check(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 || findings[0].Code != "incomplete-journal-line" {
		t.Fatalf("findings = %#v", findings)
	}
}
```

Add import:

```go
"os"
```

- [ ] **Step 2: Implement doctor checks**

Create `internal/restore/doctor.go`:

```go
package restore

import (
	"context"
	"os"

	"github.com/Diaszano/bearm/internal/journal"
)

// Finding is one non-destructive doctor diagnostic.
type Finding struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// Doctor inspects journal and filesystem consistency.
type Doctor struct {
	repository *journal.Repository
}

// NewDoctor creates a diagnostic service.
func NewDoctor(repository *journal.Repository) *Doctor {
	return &Doctor{repository: repository}
}

// Check returns diagnostics without mutating journal or trash.
func (d *Doctor) Check(ctx context.Context) ([]Finding, error) {
	readResult, err := d.repository.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	findings := make([]Finding, 0)
	if readResult.IncompleteTrailingLine {
		findings = append(findings, Finding{
			Code:    "incomplete-journal-line",
			Message: "journal contains an incomplete trailing line",
			Path:    d.repository.Path(),
		})
	}

	active, err := d.repository.ActiveItems(ctx)
	if err != nil {
		return nil, err
	}
	for _, record := range active {
		if _, err := os.Lstat(record.TrashedPath); os.IsNotExist(err) {
			findings = append(findings, Finding{
				Code:    "missing-trash-item",
				Message: "active journal record points to a missing trash item",
				Path:    record.TrashedPath,
			})
		}
	}

	return findings, nil
}
```

- [ ] **Step 3: Run doctor tests**

Run:

```bash
gofmt -w internal/restore
go test ./internal/restore -run TestDoctor -v
```

Expected:

```text
PASS
```

- [ ] **Step 4: Commit**

```bash
git add internal/restore/doctor.go internal/restore/doctor_test.go
git commit -m "feat: diagnose Bearm journal consistency"
```

### Task 6: Wire Native List, Restore, Purge, and Doctor Commands

**Files:**
- Modify: `internal/app/dependencies.go`
- Modify: `internal/app/app.go`
- Modify: `internal/app/app_test.go`
- Modify: `cmd/bearm/main.go`

**Interfaces:**
- Consumes:
  - parsed `cli.NativeRequest`;
  - `journal.Repository`;
  - restore, purge, and doctor services.
- Produces:
  - functional native commands;
  - JSON and human-readable list/doctor rendering;
  - purge confirmation through stdin.

- [ ] **Step 1: Extend application dependencies**

Replace `internal/app/dependencies.go` with:

```go
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
```

- [ ] **Step 2: Add native integration tests**

Add to `internal/app/app_test.go`:

```go
func TestRunNativeListJSON(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "journal.jsonl")
	repository := journal.New(path, time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID:        "item-1",
		OperationID:   "operation-1",
		OriginalPath:  "/work/file",
		TrashedPath:   "/trash/file",
		Backend:       "fake",
		DeletedAt:     time.Now(),
		Status:        "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{Repository: repository},
	)

	code := instance.Run(context.Background(), []string{"bearm", "list", "--json"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"item_id":"item-1"`) {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunNativeRestoreLast(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	trashed := filepath.Join(root, "trash", "file")
	original := filepath.Join(root, "work", "file")
	if err := os.MkdirAll(filepath.Dir(trashed), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(trashed, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	repository := journal.New(filepath.Join(root, "journal.jsonl"), time.Now)
	record := domain.TrashRecord{
		SchemaVersion: 1,
		ItemID: "item-1", OperationID: "operation-1",
		OriginalPath: original, TrashedPath: trashed,
		Backend: "fake", DeletedAt: time.Now(), Status: "trashed",
	}
	if err := repository.Append(context.Background(), []domain.TrashRecord{record}); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	instance := app.NewWithDependencies(
		strings.NewReader(""),
		&stdout,
		&stderr,
		buildinfo.Current(),
		app.Dependencies{Repository: repository},
	)

	code := instance.Run(context.Background(), []string{"bearm", "restore", "--last"})
	if code != 0 {
		t.Fatalf("Run() code = %d, stderr = %q", code, stderr.String())
	}
	if _, err := os.Stat(original); err != nil {
		t.Fatalf("restored path missing: %v", err)
	}
}
```

Add imports:

```go
"time"

"github.com/Diaszano/bearm/internal/domain"
"github.com/Diaszano/bearm/internal/journal"
```

- [ ] **Step 3: Implement native command dispatch**

In `internal/app/app.go`, change `runNative` to accept `ctx`:

```go
func (a *App) runNative(ctx context.Context, args []string) int
```

Pass `ctx` from `Run`.

After parsing the request and handling `version`, require repository infrastructure:

```go
if a.dependencies.Repository == nil {
	fmt.Fprintln(a.err, "bearm: repositório de histórico não configurado")
	return 1
}
```

Add a command switch:

```go
switch request.Command {
case cli.CommandList:
	return a.runList(ctx, request)
case cli.CommandRestore:
	return a.runRestore(ctx, request)
case cli.CommandPurge:
	return a.runPurge(ctx, request)
case cli.CommandDoctor:
	return a.runDoctor(ctx, request)
default:
	fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
	return 2
}
```

Add these methods:

```go
func (a *App) runList(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.dependencies.Repository.ActiveItems(ctx)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}
	if len(records) > request.Limit {
		records = records[:request.Limit]
	}

	if request.JSON {
		encoder := json.NewEncoder(a.out)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(records); err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
		return 0
	}

	for _, record := range records {
		fmt.Fprintf(a.out, "%s\t%s\t%s\n", record.ItemID, record.DeletedAt.Local().Format(time.RFC3339), record.OriginalPath)
	}
	return 0
}

func (a *App) selectRecords(ctx context.Context, request cli.NativeRequest) ([]domain.TrashRecord, error) {
	if request.Last {
		return a.dependencies.Repository.LatestOperation(ctx)
	}
	if request.Operation != "" {
		active, err := a.dependencies.Repository.ActiveItems(ctx)
		if err != nil {
			return nil, err
		}
		records := make([]domain.TrashRecord, 0)
		for _, record := range active {
			if record.OperationID == request.Operation {
				records = append(records, record)
			}
		}
		if len(records) == 0 {
			return nil, errors.New("operação ativa não encontrada")
		}
		return records, nil
	}
	return a.dependencies.Repository.FindItems(ctx, request.ItemIDs)
}

func (a *App) runRestore(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.selectRecords(ctx, request)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	service := restore.NewService(a.dependencies.Repository, time.Now)
	results := service.Restore(ctx, records, restore.CollisionFail)
	return renderNativeResults(a.out, a.err, results)
}

func (a *App) runPurge(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.selectRecords(ctx, request)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	confirmed := request.Yes
	if !confirmed {
		prompter := removal.NewPrompter(a.stdin, a.err)
		confirmed, err = prompter.ConfirmTarget("itens selecionados permanentemente")
		if err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
	}

	purger := restore.NewPurger(a.dependencies.Repository, time.Now)
	results := purger.Purge(ctx, records, confirmed)
	return renderNativeResults(a.out, a.err, results)
}

func (a *App) runDoctor(ctx context.Context, request cli.NativeRequest) int {
	findings, err := restore.NewDoctor(a.dependencies.Repository).Check(ctx)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	if request.JSON {
		if err := json.NewEncoder(a.out).Encode(findings); err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
	} else if len(findings) == 0 {
		fmt.Fprintln(a.out, "Nenhum problema encontrado.")
	} else {
		for _, finding := range findings {
			fmt.Fprintf(a.out, "%s: %s (%s)\n", finding.Code, finding.Message, finding.Path)
		}
	}

	if len(findings) > 0 {
		return 1
	}
	return 0
}

func renderNativeResults(stdout, stderr io.Writer, results []domain.ItemResult) int {
	failed := false
	for _, result := range results {
		if result.Status == domain.ItemFailed {
			failed = true
			fmt.Fprintf(stderr, "bearm: %s: %v\n", result.Path, result.Err)
			continue
		}
		fmt.Fprintf(stdout, "%s\t%s\n", result.Status, result.Path)
	}
	if failed {
		return 1
	}
	return 0
}
```

Add imports:

```go
"encoding/json"
"errors"
"time"

"github.com/Diaszano/bearm/internal/restore"
```

- [ ] **Step 4: Assemble the real repository in `main`**

In `cmd/bearm/main.go`, resolve home and platform directories, then create dependencies:

```go
home, err := os.UserHomeDir()
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
dirs, err := platform.ResolveDirs(os.Getenv, home, runtime.GOOS)
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

repository := journal.New(filepath.Join(dirs.StateRoot, "journal.jsonl"), time.Now)
policy, err := safety.NewPolicy(safety.Config{
	HardProtectedRoots: []string{dirs.ConfigRoot, dirs.StateRoot, dirs.DataRoot},
})
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

dependencies := app.Dependencies{
	Backend:    newPlatformBackend(home),
	Journal:    repository,
	Repository: repository,
	Policy:     policy,
}
instance := app.NewWithDependencies(
	os.Stdin,
	os.Stdout,
	os.Stderr,
	buildinfo.Current(),
	dependencies,
)
```

Move platform backend construction into an exported application function:

```go
// NewPlatformBackend creates the host trash backend.
func NewPlatformBackend(home string) domain.TrashBackend {
	return newPlatformBackend(home)
}
```

Use `app.NewPlatformBackend(home)` from `main`.

Add imports:

```go
"fmt"
"path/filepath"
"runtime"
"time"

"github.com/Diaszano/bearm/internal/journal"
"github.com/Diaszano/bearm/internal/platform"
"github.com/Diaszano/bearm/internal/safety"
```

- [ ] **Step 5: Run native command tests**

Run:

```bash
gofmt -w cmd/bearm internal/app
go test ./internal/app ./internal/journal ./internal/restore -v
go test -race ./internal/app ./internal/journal ./internal/restore
```

Expected:

```text
PASS
```

- [ ] **Step 6: Manually verify isolated native flow**

Run:

```bash
tmp="$(mktemp -d)"
export HOME="$tmp/home"
export BEARM_STATE_HOME="$tmp/state"
mkdir -p "$HOME" "$tmp/work"
printf 'data' > "$tmp/work/file.txt"
go run ./cmd/bearm rm "$tmp/work/file.txt"
go run ./cmd/bearm list
go run ./cmd/bearm restore --last
test -f "$tmp/work/file.txt"
```

Expected:

```text
The list command shows one item.
The restore command reports a restored item.
The original file exists again.
```

- [ ] **Step 7: Commit**

```bash
git add cmd/bearm internal/app
git commit -m "feat: add Bearm recovery commands"
```

## Plan Completion Verification

Run:

```bash
make verify
```

Then run the isolated remove/list/restore flow and an isolated remove/purge flow.

Expected:

```text
Journal lines remain valid JSON under concurrent writes.
Restore returns items to their original locations.
Purge never runs without explicit confirmation.
Doctor reports missing items and incomplete trailing journal lines.
```
