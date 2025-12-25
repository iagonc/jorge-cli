package playbook

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// RunOptions configures playbook execution
type RunOptions struct {
	DryRun       bool
	Verbose      bool
	StartFrom    string
	StopAt       string
	SkipConfirm  bool
	OutputFormat string
}

// PlaybookUsecase orchestrates playbook operations
type PlaybookUsecase struct {
	Logger    *zap.Logger
	Executor  *Executor
	Discovery *Discovery
	Builtins  *BuiltinRegistry
}

// NewPlaybookUsecase creates a new PlaybookUsecase
func NewPlaybookUsecase(logger *zap.Logger) *PlaybookUsecase {
	builtins := NewBuiltinRegistry(logger)
	return &PlaybookUsecase{
		Logger:    logger,
		Executor:  NewExecutor(logger, builtins),
		Discovery: NewDiscovery(logger),
		Builtins:  builtins,
	}
}

// LoadPlaybook loads and validates a playbook from file
func (u *PlaybookUsecase) LoadPlaybook(ctx context.Context, path string) (*models.Playbook, error) {
	u.Logger.Info("Loading playbook", zap.String("path", path))

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read playbook file: %w", err)
	}

	var playbook models.Playbook
	if err := yaml.Unmarshal(data, &playbook); err != nil {
		return nil, fmt.Errorf("failed to parse playbook: %w", err)
	}

	playbook.FilePath = path

	// Validate playbook
	if err := u.validatePlaybook(&playbook); err != nil {
		return nil, err
	}

	return &playbook, nil
}

// LoadPlaybookByName finds and loads a playbook by name
func (u *PlaybookUsecase) LoadPlaybookByName(ctx context.Context, name string) (*models.Playbook, error) {
	path, err := u.Discovery.FindByName(name)
	if err != nil {
		return nil, err
	}
	return u.LoadPlaybook(ctx, path)
}

// ValidatePlaybook validates a playbook structure
func (u *PlaybookUsecase) validatePlaybook(playbook *models.Playbook) error {
	if playbook.Name == "" {
		return fmt.Errorf("playbook name is required")
	}

	if len(playbook.Steps) == 0 {
		return fmt.Errorf("playbook must have at least one step")
	}

	// Validate each step
	for i, step := range playbook.Steps {
		if step.ID == "" {
			return fmt.Errorf("step %d is missing an ID", i+1)
		}
		if step.Type == "" {
			return fmt.Errorf("step %s is missing a type", step.ID)
		}
		if err := u.validateStep(&step); err != nil {
			return fmt.Errorf("step %s: %w", step.ID, err)
		}
	}

	return nil
}

func (u *PlaybookUsecase) validateStep(step *models.Step) error {
	switch step.Type {
	case models.StepTypeShell:
		if step.Shell == nil || step.Shell.Command == "" {
			return fmt.Errorf("shell step requires a command")
		}
	case models.StepTypeBuiltin:
		if step.Builtin == nil || step.Builtin.Function == "" {
			return fmt.Errorf("builtin step requires a function")
		}
		if _, exists := u.Builtins.Get(step.Builtin.Function); !exists {
			return fmt.Errorf("unknown builtin function: %s", step.Builtin.Function)
		}
	case models.StepTypeParallel:
		if len(step.Parallel) == 0 {
			return fmt.Errorf("parallel step requires sub-steps")
		}
	case models.StepTypeConditional:
		if step.Conditional == nil || step.Conditional.If == "" {
			return fmt.Errorf("conditional step requires an 'if' condition")
		}
	case models.StepTypeLoop:
		if step.Loop == nil {
			return fmt.Errorf("loop step requires loop configuration")
		}
	default:
		return fmt.Errorf("unknown step type: %s", step.Type)
	}
	return nil
}

// CheckRequirements verifies that required tools are installed
func (u *PlaybookUsecase) CheckRequirements(playbook *models.Playbook) ([]string, error) {
	var missing []string

	for _, tool := range playbook.Requirements.Tools {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}

	return missing, nil
}

// RunPlaybook executes a playbook with given variables
func (u *PlaybookUsecase) RunPlaybook(ctx context.Context, playbook *models.Playbook, vars map[string]string, opts RunOptions) (*models.PlaybookResult, error) {
	u.Logger.Info("Running playbook",
		zap.String("name", playbook.Name),
		zap.Bool("dryRun", opts.DryRun))

	// Check requirements
	missing, _ := u.CheckRequirements(playbook)
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required tools: %v", missing)
	}

	// Validate required inputs
	for _, input := range playbook.Inputs {
		if input.Required {
			if _, exists := vars[input.Name]; !exists {
				if input.Default != "" {
					vars[input.Name] = input.Default
				} else {
					return nil, fmt.Errorf("required variable not provided: %s", input.Name)
				}
			}
		}
	}

	return u.Executor.Execute(ctx, playbook, vars, opts)
}

// DiscoverPlaybooks finds all playbooks in configured locations
func (u *PlaybookUsecase) DiscoverPlaybooks(ctx context.Context) ([]models.PlaybookInfo, error) {
	return u.Discovery.DiscoverAll()
}
