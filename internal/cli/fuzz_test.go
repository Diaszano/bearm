package cli_test

import (
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/domain"
)

func FuzzParseCompatibility(f *testing.F) {
	for _, seed := range []string{
		"-rf build",
		"-if file",
		"--interactive=once a b c d",
		"-- -rf",
		"file -rf",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		args := strings.Fields(input)
		for _, profile := range []domain.CompatibilityProfile{
			domain.ProfileGNU,
			domain.ProfileBSD,
			domain.ProfilePOSIX,
		} {
			request, err := cli.ParseCompatibility(args, profile)
			if err == nil && request.Profile != profile {
				t.Fatalf("Profile = %q, want %q", request.Profile, profile)
			}
		}
	})
}

func FuzzParseNative(f *testing.F) {
	for _, seed := range []string{
		"list --limit 10",
		"restore --last",
		"purge --yes item-1",
		"doctor --json",
		"config check",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(_ *testing.T, input string) {
		_, _ = cli.ParseNative(strings.Fields(input))
	})
}
