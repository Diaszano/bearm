package i18n_test

import (
	"testing"

	"github.com/Diaszano/bearm/internal/i18n"
)

func TestCatalogPortugueseUnknownCommand(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(i18n.LanguagePTBR)
	if got := catalog.Text(i18n.MessageUnknownCommand); got != "comando desconhecido" {
		t.Fatalf("Text() = %q", got)
	}
}

func TestCatalogEnglishMissingOperand(t *testing.T) {
	t.Parallel()

	catalog := i18n.NewCatalog(i18n.LanguageEN)
	if got := catalog.Text(i18n.MessageMissingOperand); got != "missing operand" {
		t.Fatalf("Text() = %q", got)
	}
}
