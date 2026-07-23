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

func TestAllMessagesTranslated(t *testing.T) {
	t.Parallel()

	messages := []i18n.Message{
		i18n.MessageUnknownCommand,
		i18n.MessageMissingOperand,
		i18n.MessageIllegalOption,
		i18n.MessageUnrecognizedOption,
		i18n.MessageInvalidInteractive,
		i18n.MessageNativeUsage,
	}

	for _, lang := range []i18n.Language{i18n.LanguagePTBR, i18n.LanguageEN} {
		catalog := i18n.NewCatalog(lang)
		for _, msg := range messages {
			text := catalog.Text(msg)
			if text == "" {
				t.Errorf("language %s: message %q is empty", lang, msg)
			}
		}
	}
}
