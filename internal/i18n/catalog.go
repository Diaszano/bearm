package i18n

// Message identifies one translated message.
type Message string

const (
	// MessageUnknownCommand reports an unsupported native command.
	MessageUnknownCommand Message = "unknown_command"
	// MessageMissingOperand reports an empty compatibility operand list.
	MessageMissingOperand Message = "missing_operand"
	// MessageMissingOperandTryHelp renders the 'try --help' suggestion.
	MessageMissingOperandTryHelp Message = "missing_operand_try_help"
)

// Catalog renders typed messages.
type Catalog struct {
	messages map[Message]string
}

// NewCatalog creates a complete catalog for language.
func NewCatalog(language Language) Catalog {
	if language == LanguagePTBR {
		return Catalog{messages: map[Message]string{
			MessageUnknownCommand:        "comando desconhecido", //nolint:misspell // "comando" is Portuguese.
			MessageMissingOperand:        "operando ausente",
			MessageMissingOperandTryHelp: "Experimente '%s --help' para mais informações.",
		}}
	}

	return Catalog{messages: map[Message]string{
		MessageUnknownCommand:        "unknown command",
		MessageMissingOperand:        "missing operand",
		MessageMissingOperandTryHelp: "Try '%s --help' for more information.",
	}}
}

// Text returns the translated text for key.
func (c Catalog) Text(key Message) string {
	return c.messages[key]
}
