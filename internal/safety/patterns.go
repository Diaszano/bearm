package safety

import (
	"errors"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// PatternMatcher evaluates gitignore-compatible patterns against absolute paths.
type PatternMatcher struct {
	compiled *ignore.GitIgnore
}

// CompilePatterns compiles absolute gitignore-compatible patterns.
func CompilePatterns(lines []string) (*PatternMatcher, error) {
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "!") {
			if !strings.HasPrefix(line, "!/") {
				return nil, errors.New("protected patterns must be absolute or negated absolute patterns")
			}
			normalized = append(normalized, "!"+strings.TrimPrefix(line, "!/"))
		} else {
			if !strings.HasPrefix(line, "/") {
				return nil, errors.New("protected patterns must be absolute or negated absolute patterns")
			}
			normalized = append(normalized, strings.TrimPrefix(line, "/"))
		}
	}
	return &PatternMatcher{compiled: ignore.CompileIgnoreLines(normalized...)}, nil
}

// Matches reports whether absolutePath is protected.
func (m *PatternMatcher) Matches(absolutePath string) bool {
	if m == nil || m.compiled == nil || !filepath.IsAbs(absolutePath) {
		return false
	}
	relativeToRoot := strings.TrimPrefix(filepath.ToSlash(filepath.Clean(absolutePath)), "/")
	return m.compiled.MatchesPath(relativeToRoot)
}
