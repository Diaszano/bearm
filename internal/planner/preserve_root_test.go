package planner_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/planner"
)

func TestValidatePreserveRootAllRejectsDirectoryOnDifferentDeviceFromParent(t *testing.T) {
	t.Parallel()

	path := "/mnt/external"
	deviceID := func(current string) (uint64, error) {
		if current == path {
			return 20, nil
		}
		if current == filepath.Dir(path) {
			return 10, nil
		}
		return 10, nil
	}

	err := planner.ValidatePreserveRootAll(
		path,
		domain.TargetDir,
		domain.PreserveRootAll,
		deviceID,
	)
	if err == nil {
		t.Fatal("ValidatePreserveRootAll() error = nil, want non-nil")
	}
}

func TestValidatePreserveRootAllAllowsRegularFile(t *testing.T) {
	t.Parallel()

	err := planner.ValidatePreserveRootAll(
		"/mnt/external/file.txt",
		domain.TargetFile,
		domain.PreserveRootAll,
		func(string) (uint64, error) { return 20, nil },
	)
	if err != nil {
		t.Fatalf("ValidatePreserveRootAll() error = %v", err)
	}
}
