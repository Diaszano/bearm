package i18n

// Message identifies one translated message.
type Message string

const (
	// MessageUnknownCommand reports an unsupported native command.
	MessageUnknownCommand Message = "unknown_command"
	// MessageMissingOperand reports an empty compatibility operand list.
	MessageMissingOperand Message = "missing_operand"
	// MessageIllegalOption reports an unsupported short option.
	MessageIllegalOption Message = "illegal_option"
	// MessageUnrecognizedOption reports an unsupported long option.
	MessageUnrecognizedOption Message = "unrecognized_option"
	// MessageInvalidInteractive reports an invalid --interactive value.
	MessageInvalidInteractive Message = "invalid_interactive"
	// MessageNativeUsage is the native Bearm usage heading.
	MessageNativeUsage Message = "native_usage"
)

// Catalog renders typed messages.
type Catalog struct {
	messages map[Message]string
}

// NewCatalog creates a complete catalog for language.
func NewCatalog(language Language) Catalog {
	if language == LanguagePTBR {
		return Catalog{messages: map[Message]string{
			MessageUnknownCommand:     "comando desconhecido", //nolint:misspell // "comando" is Portuguese.
			MessageMissingOperand:     "operando ausente",
			MessageIllegalOption:      "opção ilegal",
			MessageUnrecognizedOption: "opção não reconhecida",
			MessageInvalidInteractive: "valor inválido para --interactive",
			MessageNativeUsage:        "Uso: bearm <comando> [opções]", //nolint:misspell // "comando" is Portuguese.
		}}
	}

	return Catalog{messages: map[Message]string{
		MessageUnknownCommand:     "unknown command",
		MessageMissingOperand:     "missing operand",
		MessageIllegalOption:      "illegal option",
		MessageUnrecognizedOption: "unrecognized option",
		MessageInvalidInteractive: "invalid value for --interactive",
		MessageNativeUsage:        "Usage: bearm <command> [options]",
	}}
}

// Text returns the translated text for key.
func (c Catalog) Text(key Message) string {
	return c.messages[key]
}
