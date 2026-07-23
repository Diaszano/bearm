package trash

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Reservation owns one exclusively reserved metadata name.
type Reservation struct {
	TargetPath   string
	MetadataPath string
	active       bool
}

// ReserveName reserves a destination by exclusively creating matching metadata.
func ReserveName(
	filesDir string,
	metadataDir string,
	base string,
	metadataSuffix string,
	metadata []byte,
) (Reservation, error) {
	if base == "" || base == "." || base == ".." || base == "/" || filepath.Base(base) != base {
		return Reservation{}, errors.New("invalid trash base name")
	}
	if metadataSuffix == "" || filepath.Base(metadataSuffix) != metadataSuffix {
		return Reservation{}, errors.New("invalid metadata suffix")
	}

	for suffix := 0; suffix < 10000; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", base, suffix)
		}

		targetPath := filepath.Join(filesDir, candidate)
		metadataPath := filepath.Join(metadataDir, candidate+metadataSuffix)
		if pathExists(targetPath) {
			continue
		}

		file, err := os.OpenFile(metadataPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return Reservation{}, err
		}

		writeErr := writeAndSync(file, metadata)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(metadataPath)
			return Reservation{}, writeErr
		}
		if closeErr != nil {
			_ = os.Remove(metadataPath)
			return Reservation{}, closeErr
		}
		if pathExists(targetPath) {
			_ = os.Remove(metadataPath)
			continue
		}

		return Reservation{
			TargetPath:   targetPath,
			MetadataPath: metadataPath,
			active:       true,
		}, nil
	}

	return Reservation{}, errors.New("trash name reservation limit exceeded")
}

// Commit marks the reservation as completed.
func (r *Reservation) Commit() {
	r.active = false
}

// Rollback removes metadata owned by an incomplete reservation.
func (r *Reservation) Rollback() error {
	if !r.active {
		return nil
	}
	r.active = false
	err := os.Remove(r.MetadataPath)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func pathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil || !os.IsNotExist(err)
}

func writeAndSync(file *os.File, data []byte) error {
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Sync()
}
