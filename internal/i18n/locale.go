// Package i18n owns Bearm message catalogs and locale selection.
package i18n

import "strings"

// Language identifies a supported Bearm language.
type Language string

const (
	// LanguagePTBR selects Brazilian Portuguese.
	LanguagePTBR Language = "pt-BR"
	// LanguageEN selects English.
	LanguageEN Language = "en"
)

// ResolveNativeLanguage resolves the language for Bearm-native commands.
func ResolveNativeLanguage(getenv func(string) string) Language {
	if value := normalizeLanguage(getenv("BEARM_LANG")); value != "" {
		return value
	}
	return LanguagePTBR
}

// ResolveCompatibilityLanguage resolves the locale for compatibility diagnostics.
func ResolveCompatibilityLanguage(getenv func(string) string) Language {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := normalizeLanguage(getenv(key)); value != "" {
			return value
		}
	}
	return LanguageEN
}

func normalizeLanguage(value string) Language {
	normalized := strings.ToLower(strings.ReplaceAll(value, "_", "-"))
	switch {
	case strings.HasPrefix(normalized, "pt-br"), normalized == "pt":
		return LanguagePTBR
	case strings.HasPrefix(normalized, "en"):
		return LanguageEN
	default:
		return ""
	}
}
