package i18n_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/i18n"
)

func TestResolveNativeLanguageDefaultsToBrazilianPortuguese(t *testing.T) {
	t.Parallel()

	got := i18n.ResolveNativeLanguage(func(string) string { return "" })
	if got != i18n.LanguagePTBR {
		t.Fatalf("language = %q", got)
	}
}

func TestResolveNativeLanguageHonorsBearmLanguage(t *testing.T) {
	t.Parallel()

	env := map[string]string{"BEARM_LANG": "en"}
	got := i18n.ResolveNativeLanguage(func(key string) string { return env[key] })
	if got != i18n.LanguageEN {
		t.Fatalf("language = %q", got)
	}
}

func TestResolveCompatibilityLanguageUsesLocalePrecedence(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"LANG":        "en_US.UTF-8",
		"LC_MESSAGES": "pt_BR.UTF-8",
		"LC_ALL":      "",
	}
	got := i18n.ResolveCompatibilityLanguage(func(key string) string { return env[key] })
	if got != i18n.LanguagePTBR {
		t.Fatalf("language = %q", got)
	}
}
