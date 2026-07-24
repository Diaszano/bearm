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
	prompter := removal.NewPrompter(strings.NewReader("yes\n"), &output)

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
	if output.String() == "" {
		t.Fatal("prompt output is empty")
	}
}

func TestConfirmTargetDefaultsToNo(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	prompter := removal.NewPrompter(strings.NewReader("\n"), &output)

	accepted, err := prompter.ConfirmTarget("file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if accepted {
		t.Fatal("accepted = true, want false")
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
			name: "four targets",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				Targets: make([]domain.PlannedTarget, 4),
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
				Targets: []domain.PlannedTarget{{Kind: domain.TargetDir}},
			},
			want: true,
		},
		{
			name: "three files",
			plan: domain.RemovalPlan{
				Request: domain.RemoveRequest{
					Options: domain.RemoveOptions{Interactive: domain.InteractiveOnce},
				},
				Targets: make([]domain.PlannedTarget, 3),
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
