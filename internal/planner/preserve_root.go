package planner

import (
	"errors"
	"path/filepath"

	"github.com/Diaszano/bearm/internal/domain"
)

// DeviceIDFunc returns a filesystem device ID for a path.
type DeviceIDFunc func(string) (uint64, error)

// ValidatePreserveRootAll rejects directory operands whose parent is on another device.
func ValidatePreserveRootAll(
	path string,
	kind domain.TargetKind,
	mode domain.PreserveRootMode,
	deviceID DeviceIDFunc,
) error {
	if mode != domain.PreserveRootAll || kind != domain.TargetDir {
		return nil
	}

	targetDevice, err := deviceID(path)
	if err != nil {
		return err
	}
	parentDevice, err := deviceID(filepath.Dir(path))
	if err != nil {
		return err
	}
	if targetDevice != parentDevice {
		return errors.New("preserve-root=all protects this directory operand")
	}
	return nil
}
