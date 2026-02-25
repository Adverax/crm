package procedure

import (
	"time"
)

// ExecutionResult holds the output of a procedure execution.
type ExecutionResult struct {
	Success  bool               `json:"success"`
	Result   map[string]any     `json:"result,omitempty"`
	Warnings []ExecutionWarning `json:"warnings,omitempty"`
	Trace    []TraceEntry       `json:"trace,omitempty"`
}

// ExecutionWarning is a non-fatal issue encountered during execution.
type ExecutionWarning struct {
	Command string `json:"command"`
	Message string `json:"message"`
}

// TraceEntry records a single command execution for debugging.
type TraceEntry struct {
	Step     string `json:"step"`
	Type     string `json:"type"`
	Status   string `json:"status"` // "ok", "skipped", "error", "warning", "retry"
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// ExecutionContext holds runtime state during procedure execution.
type ExecutionContext struct {
	Vars          map[string]any
	CallStack     []string
	RollbackStack []RollbackEntry
	Warnings      []ExecutionWarning
	CommandCount  int
	HTTPCount     int
	NotifyCount   int
	DryRun        bool
	Deadline      time.Time
	Trace         []TraceEntry
}

// NewExecutionContext creates a fresh ExecutionContext with initial variables.
func NewExecutionContext(input map[string]any, dryRun bool, deadline time.Time) *ExecutionContext {
	vars := map[string]any{
		"input": input,
		"now":   time.Now().UTC(),
	}
	return &ExecutionContext{
		Vars:     vars,
		DryRun:   dryRun,
		Deadline: deadline,
		Trace:    make([]TraceEntry, 0),
	}
}

// RollbackEntry represents a single rollback action in the LIFO Saga stack.
type RollbackEntry struct {
	StepName string
	Action   func() error
}

// ExecutionError is a structured error from procedure execution.
type ExecutionError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Step    string `json:"step,omitempty"`
}

func (e *ExecutionError) Error() string {
	if e.Step != "" {
		return e.Step + ": " + e.Message
	}
	return e.Message
}

// ExecOption configures execution behavior.
type ExecOption func(*execOptions)

type execOptions struct {
	dryRun bool
}

// WithDryRun enables dry-run mode.
func WithDryRun() ExecOption {
	return func(o *execOptions) {
		o.dryRun = true
	}
}
