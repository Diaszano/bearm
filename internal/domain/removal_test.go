package domain_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
)

func TestRemoveRequestValidateAcceptsGNURequest(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile:  domain.ProfileGNU,
		Operands: []string{"file.txt"},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRemoveRequestValidateRejectsUnknownProfile(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile:  domain.CompatibilityProfile("unknown"),
		Operands: []string{"file.txt"},
	}

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestRemoveRequestValidateRejectsMissingOperandsWithoutForce(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{Profile: domain.ProfileGNU}

	if err := request.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil")
	}
}

func TestRemoveRequestValidateAcceptsMissingOperandsWithForce(t *testing.T) {
	t.Parallel()

	request := domain.RemoveRequest{
		Profile: domain.ProfileGNU,
		Options: domain.RemoveOptions{Force: true},
	}

	if err := request.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
