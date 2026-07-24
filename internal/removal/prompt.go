// Package removal executes validated Bearm removal plans.
package removal

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/Diaszano/bearm/internal/domain"
)

// Prompter handles compatibility confirmation input and output.
type Prompter struct {
	reader *bufio.Reader
	writer io.Writer
}

// NewPrompter creates a confirmation prompter.
func NewPrompter(reader io.Reader, writer io.Writer) *Prompter {
	return &Prompter{reader: bufio.NewReader(reader), writer: writer}
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

// ConfirmOnce asks for operation-level confirmation.
func (p *Prompter) ConfirmOnce(plan domain.RemovalPlan) (bool, error) {
	fmt.Fprint(p.writer, "rm: remove all arguments? ")
	return p.readYes()
}

// ConfirmTarget asks for one target confirmation.
func (p *Prompter) ConfirmTarget(path string) (bool, error) {
	fmt.Fprintf(p.writer, "rm: remove %s? ", path)
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
