// Package removal executes validated Bearm removal plans.
package removal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
)

// PromptFormatter formats confirmation prompts.
type PromptFormatter interface {
	PromptOnce() string
	PromptTarget(path string) string
}

type defaultPromptFormatter struct{}

func (defaultPromptFormatter) PromptOnce() string {
	return "rm: remove all arguments? "
}

func (defaultPromptFormatter) PromptTarget(path string) string {
	return fmt.Sprintf("rm: remove %s? ", path)
}

// Prompter handles compatibility confirmation input and output.
type Prompter struct {
	reader    *bufio.Reader
	writer    io.Writer
	formatter PromptFormatter
}

// NewPrompter creates a confirmation prompter.
func NewPrompter(reader io.Reader, writer io.Writer, formatter ...PromptFormatter) *Prompter {
	var f PromptFormatter = defaultPromptFormatter{}
	if len(formatter) > 0 && formatter[0] != nil {
		f = formatter[0]
	}
	return &Prompter{reader: bufio.NewReader(reader), writer: writer, formatter: f}
}

// NeedsOncePrompt reports whether -I requires one confirmation.
func NeedsOncePrompt(plan domain.RemovalPlan) bool {
	if plan.Request.Options.Interactive != domain.InteractiveOnce {
		return false
	}
	if len(plan.Targets) > 3 {
		return true
	}
	return plan.Request.Options.Recursive
}

func (p *Prompter) getFormatter() PromptFormatter {
	if p.formatter == nil {
		return defaultPromptFormatter{}
	}
	return p.formatter
}

// ConfirmOnce asks for operation-level confirmation.
func (p *Prompter) ConfirmOnce(_ domain.RemovalPlan) (bool, error) {
	if _, err := fmt.Fprint(p.writer, p.getFormatter().PromptOnce()); err != nil {
		return false, err
	}
	return p.readYes()
}

// ConfirmTarget asks for one target confirmation.
func (p *Prompter) ConfirmTarget(path string) (bool, error) {
	if _, err := fmt.Fprint(p.writer, p.getFormatter().PromptTarget(path)); err != nil {
		return false, err
	}
	return p.readYes()
}

func (p *Prompter) readYes() (bool, error) {
	answer, err := p.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	answer = strings.TrimSpace(answer)
	return len(answer) > 0 && (answer[0] == 'y' || answer[0] == 'Y'), nil
}
