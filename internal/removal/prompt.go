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
	PromptOnce(domain.RemovalPlan) string
	PromptTarget(path string) string
}

// Prompter handles compatibility confirmation input and output.
type Prompter struct {
	reader    *bufio.Reader
	writer    io.Writer
	formatter PromptFormatter
}

// NewPrompter creates a confirmation prompter.
func NewPrompter(reader io.Reader, writer io.Writer, formatter PromptFormatter) *Prompter {
	return &Prompter{reader: bufio.NewReader(reader), writer: writer, formatter: formatter}
}

// NeedsOncePrompt reports whether -I requires one confirmation.
func NeedsOncePrompt(plan domain.RemovalPlan) bool {
	if plan.Request.Options.Interactive != domain.InteractiveOnce {
		return false
	}
	if len(plan.ValidatedOperands) == 0 {
		return false
	}
	if len(plan.ValidatedOperands) > 3 {
		return true
	}
	return plan.Request.Options.Recursive
}

// ConfirmOnce asks for operation-level confirmation.
func (p *Prompter) ConfirmOnce(plan domain.RemovalPlan) (bool, error) {
	msg := "rm: remove all arguments? "
	if p.formatter != nil {
		msg = p.formatter.PromptOnce(plan)
	}
	if _, err := fmt.Fprint(p.writer, msg); err != nil {
		return false, err
	}
	return p.readYes()
}

// ConfirmTarget asks for one target confirmation.
func (p *Prompter) ConfirmTarget(path string) (bool, error) {
	msg := fmt.Sprintf("rm: remove %s? ", path)
	if p.formatter != nil {
		msg = p.formatter.PromptTarget(path)
	}
	if _, err := fmt.Fprint(p.writer, msg); err != nil {
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
