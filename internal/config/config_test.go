package config_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/config"
)

func TestDefaultConfiguration(t *testing.T) {
	t.Parallel()

	got := config.Default()
	if got.Language != "pt-BR" {
		t.Fatalf("Language = %q", got.Language)
	}
	if got.CompatibilityProfile != "auto" {
		t.Fatalf("CompatibilityProfile = %q", got.CompatibilityProfile)
	}
	if !got.Trash.PerMount {
		t.Fatal("Trash.PerMount = false, want true")
	}
	if !got.Safety.PreserveRoot {
		t.Fatal("Safety.PreserveRoot = false, want true")
	}
	if got.Safety.InspectDescendants {
		t.Fatal("Safety.InspectDescendants = true, want false")
	}
	if got.Restore.CollisionPolicy != "fail" {
		t.Fatalf("CollisionPolicy = %q", got.Restore.CollisionPolicy)
	}
}

func TestValidateRejectsUnknownLanguage(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Language = "es"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestValidateRejectsRelativeCustomTrash(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Trash.CustomPath = "relative/trash"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestValidateRejectsInvalidCollisionPolicy(t *testing.T) {
	t.Parallel()

	value := config.Default()
	value.Restore.CollisionPolicy = "replace"
	if err := value.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}
