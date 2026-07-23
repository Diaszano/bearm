//go:build linux

package linux

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Reservation owns one exclusively reserved metadata name.
type Reservation struct {
	TargetPath string
	InfoPath   string
	active     bool
}

const maxReserveAttempts = 10000

// Reserve atomically reserves a matching files/ and info/ name.
func Reserve(root Root, base string, metadata []byte) (Reservation, error) {
	if base == "" || base == "." || base == ".." || base == "/" || filepath.Base(base) != base {
		return Reservation{}, errors.New("invalid trash base name")
	}

	for suffix := 0; suffix < maxReserveAttempts; suffix++ {
		candidate := base
		if suffix > 0 {
			candidate = fmt.Sprintf("%s.%d", base, suffix)
		}

		targetPath := filepath.Join(root.Path, "files", candidate)
		infoPath := filepath.Join(root.Path, "info", candidate+".trashinfo")

		if pathExists(targetPath) {
			continue
		}

		file, err := os.OpenFile(infoPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return Reservation{}, err
		}

		writeErr := writeAndSync(file, metadata)
		closeErr := file.Close()
		if writeErr != nil {
			_ = os.Remove(infoPath)
			return Reservation{}, writeErr
		}
		if closeErr != nil {
			_ = os.Remove(infoPath)
			return Reservation{}, closeErr
		}

		if pathExists(targetPath) {
			_ = os.Remove(infoPath)
			continue
		}

		return Reservation{
			TargetPath: targetPath,
			InfoPath:   infoPath,
			active:     true,
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
	err := os.Remove(r.InfoPath)
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
