package models

import "time"

// Playbook represents a complete playbook definition
type Playbook struct {
	Name         string               `yaml:"name" json:"name"`
	Version      string               `yaml:"version" json:"version"`
	Description  string               `yaml:"description" json:"description"`
	Author       string               `yaml:"author,omitempty" json:"author,omitempty"`
	Tags         []string             `yaml:"tags,omitempty" json:"tags,omitempty"`
	Variables    map[string]string    `yaml:"variables,omitempty" json:"variables,omitempty"`
	Inputs       []PlaybookInput      `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Requirements PlaybookRequirements `yaml:"requirements,omitempty" json:"requirements,omitempty"`
	Steps        []Step               `yaml:"steps" json:"steps"`
	Finally      []Step               `yaml:"finally,omitempty" json:"finally,omitempty"`
	FilePath     string               `yaml:"-" json:"file_path,omitempty"`
}

// PlaybookInput defines input variable validation
type PlaybookInput struct {
	Name        string `yaml:"name" json:"name"`
	Required    bool   `yaml:"required" json:"required"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string `yaml:"type,omitempty" json:"type,omitempty"` // string, integer, boolean
	Pattern     string `yaml:"pattern,omitempty" json:"pattern,omitempty"`
	Default     string `yaml:"default,omitempty" json:"default,omitempty"`
}

// PlaybookRequirements defines execution prerequisites
type PlaybookRequirements struct {
	Tools      []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	MinVersion string   `yaml:"min_version,omitempty" json:"min_version,omitempty"`
}

// StepType represents the type of step
type StepType string

const (
	StepTypeShell       StepType = "shell"
	StepTypeBuiltin     StepType = "builtin"
	StepTypeParallel    StepType = "parallel"
	StepTypeConditional StepType = "conditional"
	StepTypeLoop        StepType = "loop"
)

// Step represents a single playbook step
type Step struct {
	ID             string             `yaml:"id" json:"id"`
	Name           string             `yaml:"name" json:"name"`
	Description    string             `yaml:"description,omitempty" json:"description,omitempty"`
	Type           StepType           `yaml:"type" json:"type"`
	Shell          *ShellConfig       `yaml:"shell,omitempty" json:"shell,omitempty"`
	Builtin        *BuiltinConfig     `yaml:"builtin,omitempty" json:"builtin,omitempty"`
	Parallel       []Step             `yaml:"parallel,omitempty" json:"parallel,omitempty"`
	Conditional    *ConditionalConfig `yaml:"conditional,omitempty" json:"conditional,omitempty"`
	Loop           *LoopConfig        `yaml:"loop,omitempty" json:"loop,omitempty"`
	Condition      string             `yaml:"condition,omitempty" json:"condition,omitempty"`
	Confirm        bool               `yaml:"confirm,omitempty" json:"confirm,omitempty"`
	ConfirmMessage string             `yaml:"confirm_message,omitempty" json:"confirm_message,omitempty"`
	OnFailure      string             `yaml:"on_failure,omitempty" json:"on_failure,omitempty"` // continue, stop, goto:<id>
	Timeout        string             `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retry          *RetryConfig       `yaml:"retry,omitempty" json:"retry,omitempty"`
	Output         []OutputMapping    `yaml:"output,omitempty" json:"output,omitempty"`
}

// ShellConfig for shell command execution
type ShellConfig struct {
	Command    string            `yaml:"command" json:"command"`
	WorkingDir string            `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Shell      string            `yaml:"shell,omitempty" json:"shell,omitempty"` // bash, sh
}

// BuiltinConfig for built-in function calls
type BuiltinConfig struct {
	Function string                 `yaml:"function" json:"function"`
	Params   map[string]interface{} `yaml:"params,omitempty" json:"params,omitempty"`
}

// ConditionalConfig for conditional branching
type ConditionalConfig struct {
	If   string `yaml:"if" json:"if"`
	Then []Step `yaml:"then" json:"then"`
	Else []Step `yaml:"else,omitempty" json:"else,omitempty"`
}

// LoopConfig for iteration
type LoopConfig struct {
	Items string `yaml:"items" json:"items"` // Variable name or inline list
	As    string `yaml:"as" json:"as"`       // Loop variable name
	Steps []Step `yaml:"steps" json:"steps"`
}

// RetryConfig for retry logic
type RetryConfig struct {
	Count   int    `yaml:"count" json:"count"`
	Delay   string `yaml:"delay" json:"delay"`
	Backoff string `yaml:"backoff,omitempty" json:"backoff,omitempty"` // linear, exponential
}

// OutputMapping defines how to capture step output
type OutputMapping struct {
	Name      string `yaml:"name" json:"name"`
	From      string `yaml:"from" json:"from"`                             // result.field, stdout, exit_code
	Transform string `yaml:"transform,omitempty" json:"transform,omitempty"` // equals:X, contains:X, regex:X
}

// ExecutionStatus enumeration
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusSuccess   ExecutionStatus = "success"
	StatusFailed    ExecutionStatus = "failed"
	StatusSkipped   ExecutionStatus = "skipped"
	StatusCancelled ExecutionStatus = "cancelled"
)

// PlaybookResult represents complete execution results
type PlaybookResult struct {
	Playbook    string                 `json:"playbook"`
	Status      ExecutionStatus        `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	StepResults []StepResult           `json:"step_results"`
	Variables   map[string]interface{} `json:"variables"`
	Error       string                 `json:"error,omitempty"`
}

// StepResult represents a single step's execution result
type StepResult struct {
	StepID     string                 `json:"step_id"`
	StepName   string                 `json:"step_name"`
	Status     ExecutionStatus        `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration"`
	Output     map[string]interface{} `json:"output,omitempty"`
	Stdout     string                 `json:"stdout,omitempty"`
	Stderr     string                 `json:"stderr,omitempty"`
	ExitCode   int                    `json:"exit_code,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Retries    int                    `json:"retries,omitempty"`
	Skipped    bool                   `json:"skipped,omitempty"`
	SkipReason string                 `json:"skip_reason,omitempty"`
}

// PlaybookInfo contains discovery metadata
type PlaybookInfo struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Tags        []string `json:"tags"`
}
