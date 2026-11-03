package trash

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/domain"
)

// RecordMetadataPath returns the sidecar path for a known Bearm backend record.
func RecordMetadataPath(record domain.TrashRecord) (string, error) {
	if !filepath.IsAbs(record.TrashedPath) {
		return "", errors.New("trashed path must be absolute")
	}
	base := filepath.Base(record.TrashedPath)
	if base == "." || base == string(filepath.Separator) {
		return "", errors.New("invalid trashed path")
	}
	switch record.Backend {
	case "linux-freedesktop":
		return filepath.Join(filepath.Dir(filepath.Dir(record.TrashedPath)), "info", base+".trashinfo"), nil
	case "custom-trash":
		return filepath.Join(filepath.Dir(filepath.Dir(record.TrashedPath)), "info", base+".json"), nil
	case "darwin-system-trash":
		return filepath.Join(filepath.Dir(record.TrashedPath), ".bearm-info", base+".json"), nil
	default:
		return "", errors.New("unknown trash backend")
	}
}

// RemoveRecordMetadata removes the known sidecar without following directory symlinks.
func RemoveRecordMetadata(record domain.TrashRecord) error {
	path, err := RecordMetadataPath(record)
	if err != nil {
		return err
	}
	for dir := filepath.Dir(path); dir != string(filepath.Separator); dir = filepath.Dir(dir) {
		info, err := os.Lstat(dir)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("metadata path contains a symlink directory")
		}
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
