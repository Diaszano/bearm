package trash

import (
	"encoding/json"
	"time"
)

// RenderMetadata encodes private trash metadata as one JSON line.
func RenderMetadata(path string, deletedAt time.Time) ([]byte, error) {
	data, err := json.Marshal(struct {
		SchemaVersion int       `json:"schema_version"`
		OriginalPath  string    `json:"original_path"`
		DeletedAt     time.Time `json:"deleted_at"`
	}{1, path, deletedAt.UTC()})
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
