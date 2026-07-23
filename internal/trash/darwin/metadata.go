//go:build darwin

package darwin

import (
	"bytes"
	"encoding/json"
	"time"
)

// Metadata stores Bearm restore information for a macOS trash item.
type Metadata struct {
	SchemaVersion int       `json:"schema_version"`
	OriginalPath  string    `json:"original_path"`
	DeletedAt     time.Time `json:"deleted_at"`
}

// RenderMetadata renders one newline-terminated JSON metadata document.
func RenderMetadata(metadata Metadata) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(metadata); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
