package playbook

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"github.com/iagonc/jorge-cli/cmd/cli/internal/utils"
	"go.uber.org/zap"
)

// Executor handles playbook step execution
type Executor struct {
	Logger       *zap.Logger
	Builtins     *BuiltinRegistry
	Interpolator *Interpolator
}

// NewExecutor creates a new Executor
func NewExecutor(logger *zap.Logger, builtins *BuiltinRegistry) *Executor {
	return &Executor{
		Logger:       logger,
		Builtins:     builtins,
		Interpolator: NewInterpolator(),
	}
}

// Execute runs a playbook with full orchestration
func (e *Executor) Execute(ctx context.Context, playbook *models.Playbook, vars map[string]string, opts RunOptions) (*models.PlaybookResult, error) {
	result := &models.PlaybookResult{
		Playbook:  playbook.Name,
		Status:    models.StatusRunning,
		StartTime: time.Now(),
		Variables: make(map[string]interface{}),
	}

	// Merge default variables with provided ones
	execVars := e.mergeVariables(playbook.Variables, vars)
	for k, v := range execVars {
		result.Variables[k] = v
	}

	// Track if we should skip until a specific step
	skipUntil := opts.StartFrom
	shouldRun := skipUntil == ""

	// Execute main steps
	for _, step := range playbook.Steps {
		// Check if we should start running
		if !shouldRun && step.ID == skipUntil {
			shouldRun = true
		}

		if !shouldRun {
			result.StepResults = append(result.StepResults, models.StepResult{
				StepID:     step.ID,
				StepName:   step.Name,
				Status:     models.StatusSkipped,
				SkipReason: "Skipped (before start-from)",
			})
			continue
		}

		// Check if we should stop at this step
		if opts.StopAt != "" && step.ID == opts.StopAt {
			result.StepResults = append(result.StepResults, models.StepResult{
				StepID:     step.ID,
				StepName:   step.Name,
				Status:     models.StatusSkipped,
				SkipReason: "Skipped (stop-at reached)",
			})
			break
		}

		stepResult, err := e.executeStep(ctx, &step, execVars, opts)
		result.StepResults = append(result.StepResults, *stepResult)

		// Update variables with step outputs
		if stepResult.Output != nil {
			for k, v := range stepResult.Output {
				execVars[k] = fmt.Sprintf("%v", v)
				result.Variables[k] = v
			}
		}

		if err != nil {
			onFailure := step.OnFailure
			if onFailure == "" {
				onFailure = "stop"
			}

			if onFailure == "stop" {
				result.Status = models.StatusFailed
				result.Error = err.Error()
				break
			}
			// continue or goto handled implicitly
		}
	}

	// Always run finally steps
	for _, step := range playbook.Finally {
		stepResult, _ := e.executeStep(ctx, &step, execVars, opts)
		result.StepResults = append(result.StepResults, *stepResult)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if result.Status == models.StatusRunning {
		result.Status = models.StatusSuccess
	}

	return result, nil
}

func (e *Executor) mergeVariables(defaults map[string]string, provided map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range defaults {
		result[k] = v
	}
	for k, v := range provided {
		result[k] = v
	}

	return result
}

func (e *Executor) executeStep(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions) (*models.StepResult, error) {
	result := &models.StepResult{
		StepID:    step.ID,
		StepName:  step.Name,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	// Check condition
	if step.Condition != "" {
		shouldRun, err := e.evaluateCondition(step.Condition, vars)
		if err != nil {
			result.Status = models.StatusFailed
			result.Error = fmt.Sprintf("Condition evaluation error: %v", err)
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, err
		}
		if !shouldRun {
			result.Status = models.StatusSkipped
			result.SkipReason = "Condition not met"
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, nil
		}
	}

	// Handle confirmation
	if step.Confirm && !opts.SkipConfirm && !opts.DryRun {
		msg := step.ConfirmMessage
		if msg == "" {
			msg = fmt.Sprintf("Execute step '%s'?", step.Name)
		}
		if !utils.ConfirmAction(msg + " (yes/no): ") {
			result.Status = models.StatusSkipped
			result.SkipReason = "User declined"
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)
			return result, nil
		}
	}

	// Execute based on type
	var err error
	switch step.Type {
	case models.StepTypeShell:
		err = e.executeShell(ctx, step, vars, opts, result)
	case models.StepTypeBuiltin:
		err = e.executeBuiltin(ctx, step, vars, opts, result)
	case models.StepTypeParallel:
		err = e.executeParallel(ctx, step, vars, opts, result)
	case models.StepTypeConditional:
		err = e.executeConditional(ctx, step, vars, opts, result)
	case models.StepTypeLoop:
		err = e.executeLoop(ctx, step, vars, opts, result)
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	if err != nil {
		result.Status = models.StatusFailed
		result.Error = err.Error()
	} else if result.Status == "" {
		result.Status = models.StatusSuccess
	}

	return result, err
}

func (e *Executor) executeShell(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions, result *models.StepResult) error {
	// Interpolate command
	command, err := e.Interpolator.Interpolate(step.Shell.Command, vars)
	if err != nil {
		return fmt.Errorf("interpolation error: %w", err)
	}

	if opts.DryRun {
		result.Status = models.StatusSkipped
		result.SkipReason = fmt.Sprintf("Dry run: would execute: %s", command)
		return nil
	}

	// Setup command with timeout
	shell := step.Shell.Shell
	if shell == "" {
		shell = "bash"
	}

	var cmdCtx context.Context
	var cancel context.CancelFunc
	if step.Timeout != "" {
		timeout, parseErr := time.ParseDuration(step.Timeout)
		if parseErr == nil {
			cmdCtx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		} else {
			cmdCtx = ctx
		}
	} else {
		cmdCtx = ctx
	}

	cmd := exec.CommandContext(cmdCtx, shell, "-c", command)
	if step.Shell.WorkingDir != "" {
		cmd.Dir = step.Shell.WorkingDir
	}

	// Set environment variables
	if len(step.Shell.Env) > 0 {
		for k, v := range step.Shell.Env {
			interpolatedVal, _ := e.Interpolator.Interpolate(v, vars)
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, interpolatedVal))
		}
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()

	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		return err
	}

	result.ExitCode = 0

	// Capture outputs to variables
	e.captureOutputs(step.Output, result, vars)

	return nil
}

func (e *Executor) executeBuiltin(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions, result *models.StepResult) error {
	fn, exists := e.Builtins.Get(step.Builtin.Function)
	if !exists {
		return fmt.Errorf("unknown builtin function: %s", step.Builtin.Function)
	}

	// Interpolate parameters
	params := make(map[string]interface{})
	for k, v := range step.Builtin.Params {
		if strVal, ok := v.(string); ok {
			interpolated, _ := e.Interpolator.Interpolate(strVal, vars)
			params[k] = interpolated
		} else {
			params[k] = v
		}
	}

	if opts.DryRun {
		result.Status = models.StatusSkipped
		result.SkipReason = fmt.Sprintf("Dry run: would call %s with %v", step.Builtin.Function, params)
		return nil
	}

	output, err := fn(ctx, params)
	if err != nil {
		return err
	}

	result.Output["result"] = output

	// Capture outputs
	e.captureOutputs(step.Output, result, vars)

	return nil
}

func (e *Executor) executeParallel(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions, result *models.StepResult) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	subResults := make([]models.StepResult, len(step.Parallel))

	for i, subStep := range step.Parallel {
		wg.Add(1)
		go func(idx int, s models.Step) {
			defer wg.Done()
			subResult, err := e.executeStep(ctx, &s, vars, opts)
			mu.Lock()
			subResults[idx] = *subResult
			if err != nil && firstError == nil {
				firstError = err
			}
			mu.Unlock()
		}(i, subStep)
	}

	wg.Wait()

	// Merge outputs from parallel steps
	for _, sr := range subResults {
		for k, v := range sr.Output {
			result.Output[k] = v
		}
	}

	return firstError
}

func (e *Executor) executeConditional(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions, result *models.StepResult) error {
	shouldRunThen, err := e.evaluateCondition(step.Conditional.If, vars)
	if err != nil {
		return err
	}

	var stepsToRun []models.Step
	if shouldRunThen {
		stepsToRun = step.Conditional.Then
	} else {
		stepsToRun = step.Conditional.Else
	}

	for _, subStep := range stepsToRun {
		subResult, err := e.executeStep(ctx, &subStep, vars, opts)
		if err != nil {
			return err
		}
		// Merge outputs
		for k, v := range subResult.Output {
			result.Output[k] = v
		}
	}

	return nil
}

func (e *Executor) executeLoop(ctx context.Context, step *models.Step, vars map[string]string, opts RunOptions, result *models.StepResult) error {
	// Get items to iterate
	itemsStr, ok := vars[step.Loop.Items]
	if !ok {
		return fmt.Errorf("loop items variable not found: %s", step.Loop.Items)
	}

	items := strings.Split(itemsStr, ",")

	for _, item := range items {
		// Create a copy of vars with loop variable
		loopVars := make(map[string]string)
		for k, v := range vars {
			loopVars[k] = v
		}
		loopVars[step.Loop.As] = strings.TrimSpace(item)

		for _, subStep := range step.Loop.Steps {
			subResult, err := e.executeStep(ctx, &subStep, loopVars, opts)
			if err != nil {
				return err
			}
			for k, v := range subResult.Output {
				result.Output[k] = v
			}
		}
	}

	return nil
}

func (e *Executor) evaluateCondition(condition string, vars map[string]string) (bool, error) {
	// Simple condition evaluation
	// Supports: {{ .var == "value" }}, {{ .var != "" }}, {{ .var }}
	interpolated, err := e.Interpolator.Interpolate(condition, vars)
	if err != nil {
		return false, err
	}

	// Clean up whitespace
	interpolated = strings.TrimSpace(interpolated)

	// Check for common patterns
	if interpolated == "" || interpolated == "false" || interpolated == "0" {
		return false, nil
	}
	if interpolated == "true" || interpolated == "1" {
		return true, nil
	}

	// Default to true for non-empty values
	return true, nil
}

func (e *Executor) captureOutputs(outputs []models.OutputMapping, result *models.StepResult, vars map[string]string) {
	for _, output := range outputs {
		var value interface{}

		switch output.From {
		case "stdout":
			value = strings.TrimSpace(result.Stdout)
		case "stderr":
			value = strings.TrimSpace(result.Stderr)
		case "exit_code":
			value = result.ExitCode
		default:
			if strings.HasPrefix(output.From, "result.") {
				field := strings.TrimPrefix(output.From, "result.")
				if resultMap, ok := result.Output["result"].(map[string]interface{}); ok {
					value = resultMap[field]
				}
			}
		}

		if value != nil {
			result.Output[output.Name] = value
			vars[output.Name] = fmt.Sprintf("%v", value)
		}
	}
}
