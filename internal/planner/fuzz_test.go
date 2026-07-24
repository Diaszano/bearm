package planner_test

import (
	"path/filepath"
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func FuzzSafetyPolicyCheck(f *testing.F) {
	for _, seed := range []string{
		"/",
		"/tmp/file",
		"/tmp/../etc",
		"/Users/dias/project/link",
		"/home/dias/a b",
	} {
		f.Add(seed)
	}

	policy, err := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{"/tmp/bearm-state"},
	})
	if err != nil {
		f.Fatal(err)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if !filepath.IsAbs(input) {
			return
		}
		_ = policy.Check(input)
	})
}
