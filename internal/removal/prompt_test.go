package removal_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/removal"
)

func TestConfirmOnceAcceptsAnswerStartingWithY(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("yes\n"), &output, nil)

	accepted, err := prompter.ConfirmOnce(domain.RemovalPlan{
		Request: domain.RemoveRequest{
			Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
		},
		Targets: []domain.PlannedTarget{{InputPath: "a"}, {InputPath: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !accepted {
		t.Fatal("accepted = false, want true")
	}
	if output.String() != "rm: remove all arguments? " {
		t.Fatalf("prompt output = %q, want %q", output.String(), "rm: remove all arguments? ")
	}
}

func TestConfirmTargetDefaultsToNo(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("\n"), &output, nil)

	accepted, err := prompter.ConfirmTarget("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if accepted {
		t.Fatal("accepted = true, want false")
	}
	if output.String() != "rm: remove file.txt? " {
		t.Fatalf("prompt output = %q, want %q", output.String(), "rm: remove file.txt? ")
	}
}

type testCustomFormatter struct{}

func (testCustomFormatter) PromptOnce(domain.RemovalPlan) string {
	return "custom once? "
}

func (testCustomFormatter) PromptTarget(path string) string {
	return "custom target " + path + "? "
}

func TestConfirmWithCustomFormatter(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("y\ny\n"), &output, testCustomFormatter{})

	accepted, err := prompter.ConfirmOnce(domain.RemovalPlan{})
	if err != nil || !accepted {
		t.Fatalf("ConfirmOnce() = (%v, %v), want (true, nil)", accepted, err)
	}
	if output.String() != "custom once? " {
		t.Fatalf("output = %q, want %q", output.String(), "custom once? ")
	}

	output.Reset()
	accepted, err = prompter.ConfirmTarget("file.txt")
	if err != nil || !accepted {
		t.Fatalf("ConfirmTarget() = (%v, %v), want (true, nil)", accepted, err)
	}
	if output.String() != "custom target file.txt? " {
		t.Fatalf("output = %q, want %q", output.String(), "custom target file.txt? ")
	}
}

func TestNeedsOncePrompt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		plan domain.RemovalPlan
		want bool
	}{
		{
			name: "four validated operands",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				ValidatedOperands: make([]domain.ValidatedOperand, 4),
			},
			want: true,
		},
		{
			name: "recursive target",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{
						Interactive: domain.InteractiveOnce,
						Recursive:   true,
					},
				},
				ValidatedOperands: []domain.ValidatedOperand{{Kind: domain.TargetDir}},
			},
			want: true,
		},
		{
			name: "three files",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				ValidatedOperands: make([]domain.ValidatedOperand, 3),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := removal.NeedsOncePrompt(tt.plan); got != tt.want {
				t.Fatalf("NeedsOncePrompt() = %v, want %v", got, tt.want)
			}
		})
	}
}
