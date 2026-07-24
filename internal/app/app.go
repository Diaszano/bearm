// Package app coordinates Bearm use cases.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Diaszano/bearm/internal/buildinfo"
	"github.com/Diaszano/bearm/internal/cli"
	"github.com/Diaszano/bearm/internal/config"
	"github.com/Diaszano/bearm/internal/domain"
	"github.com/Diaszano/bearm/internal/i18n"
	"github.com/Diaszano/bearm/internal/id"
	"github.com/Diaszano/bearm/internal/journal"
	"github.com/Diaszano/bearm/internal/planner"
	"github.com/Diaszano/bearm/internal/platform"
	"github.com/Diaszano/bearm/internal/removal"
	"github.com/Diaszano/bearm/internal/restore"
	"github.com/Diaszano/bearm/internal/safety"
)

// App is the Bearm application shell.
type App struct {
	stdin        io.Reader
	out          io.Writer
	err          io.Writer
	info         buildinfo.Info
	getenv       func(string) string
	dependencies Dependencies
}

// New creates an application with default removal infrastructure.
func New(stdin io.Reader, stdout, stderr io.Writer, info buildinfo.Info) *App {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	defaultConfig := config.Default()
	backend := newPlatformBackend(home, defaultConfig)

	dirs, _ := platform.ResolveDirs(os.Getenv, home, runtime.GOOS)

	var homeTrash string
	if runtime.GOOS == "darwin" {
		homeTrash = filepath.Join(home, ".Trash")
	} else {
		dataHome := os.Getenv("XDG_DATA_HOME")
		if !filepath.IsAbs(dataHome) {
			dataHome = filepath.Join(home, ".local", "share")
		}
		homeTrash = filepath.Join(dataHome, "Trash")
	}

	policy, _ := safety.NewPolicy(safety.Config{
		HardProtectedRoots: []string{
			dirs.ConfigRoot,
			dirs.StateRoot,
			homeTrash,
		},
	})
	journalRepo := journal.New(filepath.Join(dirs.StateRoot, "journal.jsonl"), time.Now)
	return NewWithDependencies(stdin, stdout, stderr, info, Dependencies{
		Backend:    backend,
		Journal:    journalRepo,
		Repository: journalRepo,
		Policy:     policy,
		Config:     defaultConfig,
		ConfigPath: filepath.Join(dirs.ConfigRoot, "config.toml"),
	})
}

type discardJournal struct{}

func (d *discardJournal) Append(_ context.Context, _ []domain.TrashRecord) error {
	return nil
}

// NewWithDependencies creates an application with explicit infrastructure.
func NewWithDependencies(
	stdin io.Reader,
	stdout, stderr io.Writer,
	info buildinfo.Info,
	dependencies Dependencies,
) *App {
	return &App{
		stdin:        stdin,
		out:          stdout,
		err:          stderr,
		info:         info,
		getenv:       os.Getenv,
		dependencies: dependencies,
	}
}

// Run executes one Bearm invocation and returns a process exit code.
func (a *App) Run(ctx context.Context, argv []string) int {
	invocation, err := cli.ResolveInvocation(argv)
	if err != nil {
		fmt.Fprintln(a.err, err)
		return 2
	}

	if invocation.Mode == cli.ModeCompatibility {
		return a.runCompatibility(ctx, invocation.Args)
	}

	return a.runNative(ctx, invocation.Args)
}

func (a *App) runCompatibility(ctx context.Context, args []string) int {
	profile := resolveConfiguredProfile(
		a.dependencies.Config.CompatibilityProfile,
		a.getenv,
		runtime.GOOS,
	)
	request, err := cli.ParseCompatibility(args, profile)
	if err != nil {
		var usageErr *cli.UsageError
		if errors.As(err, &usageErr) {
			catalog := i18n.NewCatalog(i18n.ResolveCompatibilityLanguage(a.getenv))
			switch usageErr.Kind {
			case "invalid-interactive":
				fmt.Fprintf(a.err, "rm: %s '%s'\n", catalog.Text(i18n.MessageInvalidInteractive), usageErr.Value)
			case "unsupported-option":
				if strings.HasPrefix(usageErr.Option, "--") {
					fmt.Fprintf(a.err, "rm: %s '%s'\n", catalog.Text(i18n.MessageUnrecognizedOption), usageErr.Option)
				} else {
					letter := strings.TrimPrefix(usageErr.Option, "-")
					if profile == domain.ProfileBSD {
						fmt.Fprintf(a.err, "rm: %s -- %s\n", catalog.Text(i18n.MessageIllegalOption), letter)
					} else {
						fmt.Fprintf(a.err, "rm: %s -- '%s'\n", catalog.Text(i18n.MessageIllegalOption), letter)
					}
				}
			default:
				fmt.Fprintf(a.err, "rm: %v\n", err)
			}
		} else {
			fmt.Fprintf(a.err, "rm: %v\n", err)
		}
		return usageCode(profile)
	}

	if request.Options.ShowVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}
	if request.Options.ShowHelp {
		fmt.Fprintln(a.out, "Usage: rm [OPTION]... [FILE]...")
		return 0
	}

	if err := request.Validate(); err != nil {
		catalog := i18n.NewCatalog(i18n.ResolveCompatibilityLanguage(a.getenv))
		fmt.Fprintf(a.err, "rm: %s\n", catalog.Text(i18n.MessageMissingOperand))
		return usageCode(profile)
	}

	if len(request.Operands) == 0 {
		return 0
	}
	if a.dependencies.Backend == nil || a.dependencies.Journal == nil || a.dependencies.Policy == nil {
		fmt.Fprintln(a.err, "rm: removal infrastructure is not configured")
		return 1
	}

	instance := planner.New(a.dependencies.Policy, id.New)
	plan, planningFailures := instance.Plan(ctx, request)
	executor := removal.NewExecutor(
		a.dependencies.Backend,
		a.dependencies.Journal,
		a.dependencies.Policy,
		removal.NewPrompter(a.stdin, a.err),
		a.out,
	)
	result := executor.Execute(ctx, plan)
	result.Items = append(planningFailures, result.Items...)

	for _, item := range result.Items {
		if item.Status == domain.ItemFailed && item.Err != nil {
			fmt.Fprintf(a.err, "rm: %s: %v\n", item.Path, item.Err)
		}
	}
	return result.ExitCode(profile)
}

func (a *App) runNative(ctx context.Context, args []string) int {
	catalog := i18n.NewCatalog(i18n.ResolveNativeLanguage(a.getenv))
	request, err := cli.ParseNative(args)
	if err != nil {
		if errors.Is(err, cli.ErrUnknownCommand) || errors.Is(err, cli.ErrMissingCommand) {
			fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
		} else {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
		}
		return 2
	}

	if request.Command == cli.CommandVersion {
		fmt.Fprintln(a.out, a.info.String())
		return 0
	}

	if request.Command == cli.CommandConfig {
		switch request.ConfigOp {
		case "path":
			fmt.Fprintln(a.out, a.dependencies.ConfigPath)
			return 0
		case "check":
			if err := a.dependencies.Config.Validate(); err != nil {
				fmt.Fprintf(a.err, "bearm: configuração inválida: %v\n", err)
				return 3
			}
			fmt.Fprintln(a.out, "Configuração válida.")
			return 0
		default:
			fmt.Fprintln(a.err, "bearm: operação de configuração inválida")
			return 2
		}
	}

	if a.dependencies.Repository == nil {
		fmt.Fprintln(a.err, "bearm: repositório de histórico não configurado")
		return 1
	}

	switch request.Command {
	case cli.CommandList:
		return a.runList(ctx, request)
	case cli.CommandRestore:
		return a.runRestore(ctx, request)
	case cli.CommandPurge:
		return a.runPurge(ctx, request)
	case cli.CommandDoctor:
		return a.runDoctor(ctx, request)
	default:
		fmt.Fprintf(a.err, "bearm: %s\n", catalog.Text(i18n.MessageUnknownCommand))
		return 2
	}
}

func (a *App) runList(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.dependencies.Repository.ActiveItems(ctx)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}
	if len(records) > request.Limit {
		records = records[:request.Limit]
	}

	if request.JSON {
		encoder := json.NewEncoder(a.out)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(records); err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
		return 0
	}

	for _, record := range records {
		fmt.Fprintf(a.out, "%s\t%s\t%s\n", record.ItemID, record.DeletedAt.Local().Format(time.RFC3339), record.OriginalPath)
	}
	return 0
}

func (a *App) selectRecords(ctx context.Context, request cli.NativeRequest) ([]domain.TrashRecord, error) {
	if request.Last {
		return a.dependencies.Repository.LatestOperation(ctx)
	}
	if request.Operation != "" {
		active, err := a.dependencies.Repository.ActiveItems(ctx)
		if err != nil {
			return nil, err
		}
		records := make([]domain.TrashRecord, 0)
		for _, record := range active {
			if record.OperationID == request.Operation {
				records = append(records, record)
			}
		}
		if len(records) == 0 {
			return nil, errors.New("operação ativa não encontrada")
		}
		return records, nil
	}
	return a.dependencies.Repository.FindItems(ctx, request.ItemIDs)
}

func (a *App) runRestore(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.selectRecords(ctx, request)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	policy := restore.CollisionPolicy(a.dependencies.Config.Restore.CollisionPolicy)
	if policy == restore.CollisionOverwrite {
		fmt.Fprintln(a.err, "bearm: overwrite exige confirmação explícita e não é usado por restore padrão")
		return 2
	}
	service := restore.NewService(a.dependencies.Repository, time.Now)
	results := service.Restore(ctx, records, policy)
	return renderNativeResults(a.out, a.err, results)
}

func (a *App) runPurge(ctx context.Context, request cli.NativeRequest) int {
	records, err := a.selectRecords(ctx, request)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	confirmed := request.Yes
	if !confirmed {
		prompter := removal.NewPrompter(a.stdin, a.err)
		confirmed, err = prompter.ConfirmTarget("itens selecionados permanentemente")
		if err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
	}

	purger := restore.NewPurger(a.dependencies.Repository, time.Now)
	results := purger.Purge(ctx, records, confirmed)
	return renderNativeResults(a.out, a.err, results)
}

func (a *App) runDoctor(ctx context.Context, request cli.NativeRequest) int {
	findings, err := restore.NewDoctor(a.dependencies.Repository).Check(ctx)
	if err != nil {
		fmt.Fprintf(a.err, "bearm: %v\n", err)
		return 1
	}

	if request.JSON {
		if err := json.NewEncoder(a.out).Encode(findings); err != nil {
			fmt.Fprintf(a.err, "bearm: %v\n", err)
			return 1
		}
	} else if len(findings) == 0 {
		fmt.Fprintln(a.out, "Nenhum problema encontrado.")
	} else {
		for _, finding := range findings {
			fmt.Fprintf(a.out, "%s: %s (%s)\n", finding.Code, finding.Message, finding.Path)
		}
	}

	if len(findings) > 0 {
		return 1
	}
	return 0
}

func renderNativeResults(stdout, stderr io.Writer, results []domain.ItemResult) int {
	failed := false
	for _, result := range results {
		if result.Status == domain.ItemFailed {
			failed = true
			fmt.Fprintf(stderr, "bearm: %s: %v\n", result.Path, result.Err)
			continue
		}
		fmt.Fprintf(stdout, "%s\t%s\n", result.Status, result.Path)
	}
	if failed {
		return 1
	}
	return 0
}

func resolveConfiguredProfile(
	configured string,
	getenv func(string) string,
	goos string,
) domain.CompatibilityProfile {
	if configured != "" && configured != "auto" {
		switch configured {
		case "gnu":
			return domain.ProfileGNU
		case "bsd":
			return domain.ProfileBSD
		case "posix":
			return domain.ProfilePOSIX
		}
	}
	return resolveProfile(getenv, goos)
}

func resolveProfile(getenv func(string) string, goos string) domain.CompatibilityProfile {
	switch strings.ToLower(getenv("BEARM_COMPAT")) {
	case "gnu":
		return domain.ProfileGNU
	case "bsd":
		return domain.ProfileBSD
	case "posix":
		return domain.ProfilePOSIX
	}

	if goos == "darwin" {
		return domain.ProfileBSD
	}
	return domain.ProfileGNU
}

func usageCode(profile domain.CompatibilityProfile) int {
	if profile == domain.ProfileBSD {
		return 64
	}
	return 1
}
