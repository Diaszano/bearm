package safety_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/safety"
)

func TestPatternMatcherMatchesAbsoluteProtectedPath(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{
		"/Users/dias/Documents/important/**",
		"/Users/dias/Projects/**/.git",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !matcher.Matches("/Users/dias/Documents/important/report.pdf") {
		t.Fatal("important report was not matched")
	}
	if !matcher.Matches("/Users/dias/Projects/bearm/.git") {
		t.Fatal(".git path was not matched")
	}
	if matcher.Matches("/Users/dias/Downloads/file.zip") {
		t.Fatal("unprotected path was matched")
	}
}

func TestPolicyRejectsProtectedPattern(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{"/work/**/.git"})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := safety.NewPolicy(safety.Config{Patterns: matcher})
	if err != nil {
		t.Fatal(err)
	}

	if err := policy.Check("/work/bearm/.git"); err == nil {
		t.Fatal("Check() error = nil, want protected-pattern failure")
	}
}

func TestCompilePatternsValidation(t *testing.T) {
	t.Parallel()

	// Relative positive pattern should fail
	if _, err := safety.CompilePatterns([]string{"docs/*.md"}); err == nil {
		t.Fatal("CompilePatterns() error = nil, want error for relative pattern")
	}

	// Relative negated pattern should fail
	if _, err := safety.CompilePatterns([]string{"!docs/*.md"}); err == nil {
		t.Fatal("CompilePatterns() error = nil, want error for relative negated pattern")
	}

	// Absolute positive and negated patterns should pass
	if _, err := safety.CompilePatterns([]string{"/docs/*.md", "!/docs/exclude.md"}); err != nil {
		t.Fatalf("CompilePatterns() error = %v, want nil", err)
	}
}

func TestPatternMatcherMatchesIgnoresRelativePath(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{"/work/**"})
	if err != nil {
		t.Fatal(err)
	}

	if matcher.Matches("work/file.txt") {
		t.Fatal("Matches() = true, want false for relative path input")
	}
}

func TestPatternMatcherSupportsNegation(t *testing.T) {
	t.Parallel()

	matcher, err := safety.CompilePatterns([]string{
		"/work/**",
		"!/work/allowed.txt",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !matcher.Matches("/work/secret.txt") {
		t.Fatal("secret.txt should be matched (protected)")
	}
	if matcher.Matches("/work/allowed.txt") {
		t.Fatal("allowed.txt was matched, want false due to negation")
	}
}
