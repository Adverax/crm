package procedure

import (
	"context"
	"fmt"
	"testing"

	gocel "github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	celengine "github.com/adverax/crm/internal/platform/cel"
	"github.com/adverax/crm/internal/platform/metadata"
)

func newTestCELCache() *celengine.ProgramCache {
	env, _ := gocel.NewEnv(
		gocel.Variable("input", gocel.DynType),
		gocel.Variable("user", gocel.DynType),
		gocel.Variable("now", gocel.TimestampType),
		gocel.Variable("error", gocel.DynType),
		gocel.Variable("step1", gocel.DynType),
		gocel.Variable("result", gocel.DynType),
		ext.Strings(),
	)
	return celengine.NewProgramCache(env)
}

func TestEngine_ExecuteDefinition_ComputeTransform(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	eng := NewEngine(cache, nil, NewComputeCommandExecutor(NewExpressionResolver(cache)))

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type:  "compute.transform",
				As:    "step1",
				Value: map[string]string{"greeting": "hello"},
			},
		},
		Result: map[string]string{
			"message": "$.step1.greeting",
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "hello", result.Result["message"])
}

func TestEngine_ExecuteDefinition_ComputeValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		condition string
		input     map[string]any
		wantErr   bool
	}{
		{
			name:      "passes when condition is true",
			condition: "input.age > 18",
			input:     map[string]any{"age": 25},
			wantErr:   false,
		},
		{
			name:      "fails when condition is false",
			condition: "input.age > 18",
			input:     map[string]any{"age": 10},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := newTestCELCache()
			eng := NewEngine(cache, nil, NewComputeCommandExecutor(NewExpressionResolver(cache)))

			def := &metadata.ProcedureDefinition{
				Commands: []metadata.CommandDef{
					{
						Type:      "compute.validate",
						Condition: tt.condition,
						Code:      "AGE_CHECK",
						Message:   "must be over 18",
					},
				},
			}

			_, err := eng.ExecuteDefinition(context.Background(), def, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEngine_ExecuteDefinition_ComputeFail(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	eng := NewEngine(cache, nil, NewComputeCommandExecutor(NewExpressionResolver(cache)))

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type:    "compute.fail",
				Code:    "ABORT",
				Message: "intentional failure",
			},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "intentional failure")
}

func TestEngine_ExecuteDefinition_WhenCondition(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	eng := NewEngine(cache, nil, NewComputeCommandExecutor(NewExpressionResolver(cache)))

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type:  "compute.transform",
				As:    "step1",
				When:  "input.skip == false",
				Value: map[string]string{"x": "1"},
			},
		},
	}

	// skip == true → step should be skipped
	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{"skip": true})
	require.NoError(t, err)
	assert.True(t, result.Success)
	// step1 should not exist in trace with status "ok"
	for _, tr := range result.Trace {
		if tr.Step == "step1" {
			assert.Equal(t, "skipped", tr.Status)
		}
	}

	// skip == false → step should execute
	result2, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{"skip": false})
	require.NoError(t, err)
	assert.True(t, result2.Success)
}

func TestEngine_ExecuteDefinition_OptionalCommand(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	eng := NewEngine(cache, nil, NewComputeCommandExecutor(NewExpressionResolver(cache)))

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type:     "compute.fail",
				Optional: true,
				Code:     "WARN",
				Message:  "this is ok",
			},
			{
				Type:  "compute.transform",
				As:    "step1",
				Value: map[string]string{"done": "yes"},
			},
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Len(t, result.Warnings, 1)
	assert.Equal(t, "this is ok", result.Warnings[0].Message)
}

func TestEngine_ExecuteDefinition_FlowIf(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type:      "flow.if",
				Condition: "input.active == true",
				Then: []metadata.CommandDef{
					{Type: "compute.transform", As: "step1", Value: map[string]string{"branch": "then"}},
				},
				Else: []metadata.CommandDef{
					{Type: "compute.transform", As: "step1", Value: map[string]string{"branch": "else"}},
				},
			},
		},
		Result: map[string]string{
			"branch": "$.step1.branch",
		},
	}

	// active=true → then branch
	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{"active": true})
	require.NoError(t, err)
	assert.Equal(t, "then", result.Result["branch"])

	// active=false → else branch
	result2, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{"active": false})
	require.NoError(t, err)
	assert.Equal(t, "else", result2.Result["branch"])
}

func TestEngine_ExecuteDefinition_RollbackOnError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		commands        []metadata.CommandDef
		wantErr         bool
		wantErrContains string
		wantRollbackRan bool
		wantResultKey   string
		wantResultVal   any
	}{
		{
			name: "rollback runs in LIFO order when later step fails",
			commands: []metadata.CommandDef{
				{
					Type:  "compute.transform",
					As:    "step1",
					Value: map[string]string{"created": "yes"},
					Rollback: []metadata.CommandDef{
						{
							Type:  "compute.transform",
							As:    "rollback1",
							Value: map[string]string{"undone": "step1"},
						},
					},
				},
				{
					Type:    "compute.fail",
					Code:    "BOOM",
					Message: "something broke",
				},
			},
			wantErr:         true,
			wantErrContains: "something broke",
			wantRollbackRan: true,
		},
		{
			name: "rollback does not run when all steps succeed",
			commands: []metadata.CommandDef{
				{
					Type:  "compute.transform",
					As:    "step1",
					Value: map[string]string{"ok": "yes"},
					Rollback: []metadata.CommandDef{
						{
							Type:  "compute.transform",
							As:    "rollback1",
							Value: map[string]string{"undone": "step1"},
						},
					},
				},
			},
			wantErr:       false,
			wantResultKey: "val",
			wantResultVal: "yes",
		},
		{
			name: "multiple rollbacks run in reverse order",
			commands: []metadata.CommandDef{
				{
					Type:  "compute.transform",
					As:    "step1",
					Value: map[string]string{"order": "first"},
					Rollback: []metadata.CommandDef{
						{
							Type:  "compute.transform",
							As:    "rb1",
							Value: map[string]string{"rb": "1"},
						},
					},
				},
				{
					Type:  "compute.transform",
					As:    "step2",
					Value: map[string]string{"order": "second"},
					Rollback: []metadata.CommandDef{
						{
							Type:  "compute.transform",
							As:    "rb2",
							Value: map[string]string{"rb": "2"},
						},
					},
				},
				{
					Type:    "compute.fail",
					Code:    "ERR",
					Message: "fail after two steps",
				},
			},
			wantErr:         true,
			wantErrContains: "fail after two steps",
			wantRollbackRan: true,
		},
		{
			name: "rollback with multiple commands per step",
			commands: []metadata.CommandDef{
				{
					Type:  "compute.transform",
					As:    "step1",
					Value: map[string]string{"created": "yes"},
					Rollback: []metadata.CommandDef{
						{
							Type:  "compute.transform",
							As:    "rb1a",
							Value: map[string]string{"undo": "first"},
						},
						{
							Type:  "compute.transform",
							As:    "rb1b",
							Value: map[string]string{"undo": "second"},
						},
					},
				},
				{
					Type:    "compute.fail",
					Code:    "ERR",
					Message: "trigger multi-command rollback",
				},
			},
			wantErr:         true,
			wantErrContains: "trigger multi-command rollback",
			wantRollbackRan: true,
		},
		{
			name: "no rollback registered for steps without rollback field",
			commands: []metadata.CommandDef{
				{
					Type:  "compute.transform",
					As:    "step1",
					Value: map[string]string{"v": "1"},
				},
				{
					Type:    "compute.fail",
					Code:    "ERR",
					Message: "no rollback defined",
				},
			},
			wantErr:         true,
			wantErrContains: "no rollback defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cache := newTestCELCache()
			computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
			eng := NewEngine(cache, nil, computeExec)

			def := &metadata.ProcedureDefinition{
				Commands: tt.commands,
			}
			if tt.wantResultKey != "" {
				def.Result = map[string]string{
					tt.wantResultKey: "$.step1.ok",
				}
			}

			result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
				assert.True(t, result.Success)
				if tt.wantResultKey != "" {
					assert.Equal(t, tt.wantResultVal, result.Result[tt.wantResultKey])
				}
			}
		})
	}
}

func TestEngine_ExecuteDefinition_UnknownCategory(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	eng := NewEngine(cache, nil)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{Type: "unknown.cmd"},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown command category")
}

// countingExecutor is a test executor that fails N times then succeeds.
type countingExecutor struct {
	category string
	calls    int
	failFor  int // fail the first N calls
}

func (e *countingExecutor) Category() string { return e.category }

func (e *countingExecutor) Execute(_ context.Context, _ metadata.CommandDef, _ *ExecutionContext) (any, error) {
	e.calls++
	if e.calls <= e.failFor {
		return nil, fmt.Errorf("transient error (attempt %d)", e.calls)
	}
	return map[string]any{"ok": true}, nil
}

// alwaysFailExecutor always fails with a given error.
type alwaysFailExecutor struct {
	category string
	err      error
}

func (e *alwaysFailExecutor) Category() string { return e.category }

func (e *alwaysFailExecutor) Execute(_ context.Context, _ metadata.CommandDef, _ *ExecutionContext) (any, error) {
	return nil, e.err
}

func TestEngine_RetrySuccess(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	exec := &countingExecutor{category: "test", failFor: 1}
	eng := NewEngine(cache, nil, exec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "test.do",
				As:   "step1",
				Retry: &metadata.RetryConfig{
					MaxAttempts: 3,
					DelayMs:     100,
				},
			},
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, 2, exec.calls)
}

func TestEngine_RetryExhausted(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	exec := &alwaysFailExecutor{category: "test", err: fmt.Errorf("always fails")}
	eng := NewEngine(cache, nil, exec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "test.do",
				Retry: &metadata.RetryConfig{
					MaxAttempts: 3,
					DelayMs:     100,
				},
			},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "always fails")
}

func TestEngine_RetryTraceEntries(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	exec := &countingExecutor{category: "test", failFor: 2}
	eng := NewEngine(cache, nil, exec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "test.do",
				As:   "step1",
				Retry: &metadata.RetryConfig{
					MaxAttempts: 3,
					DelayMs:     100,
				},
			},
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)

	retryCount := 0
	for _, tr := range result.Trace {
		if tr.Status == "retry" {
			retryCount++
		}
	}
	assert.Equal(t, 2, retryCount)
}

func TestEngine_RetryRespectsDeadline(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	exec := &alwaysFailExecutor{category: "test", err: fmt.Errorf("fails")}
	eng := NewEngine(cache, nil, exec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "test.do",
				Retry: &metadata.RetryConfig{
					MaxAttempts: 5,
					DelayMs:     60000, // 60s delay — exceeds any reasonable deadline
				},
			},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "deadline would be exceeded")
}

func TestEngine_FlowTry_TrySucceeds(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				As:   "result",
				Try: []metadata.CommandDef{
					{Type: "compute.transform", As: "step1", Value: map[string]string{"ok": "yes"}},
				},
				Catch: []metadata.CommandDef{
					{Type: "compute.transform", As: "fallback", Value: map[string]string{"fallback": "yes"}},
				},
			},
		},
		Result: map[string]string{
			"caught": "$.result.caught",
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, false, result.Result["caught"])
}

func TestEngine_FlowTry_CatchHandlesError(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				As:   "result",
				Try: []metadata.CommandDef{
					{Type: "compute.fail", Code: "TEST_ERR", Message: "test failure"},
				},
				Catch: []metadata.CommandDef{
					{Type: "compute.transform", As: "fallback", Value: map[string]string{"recovered": "yes"}},
				},
			},
		},
		Result: map[string]string{
			"caught": "$.result.caught",
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, true, result.Result["caught"])
}

func TestEngine_FlowTry_CatchFails(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				As:   "result",
				Try: []metadata.CommandDef{
					{Type: "compute.fail", Code: "ERR1", Message: "try fails"},
				},
				Catch: []metadata.CommandDef{
					{Type: "compute.fail", Code: "ERR2", Message: "catch also fails"},
				},
			},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "catch also fails")
}

func TestEngine_FlowTry_NoCatchBlock(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				Try: []metadata.CommandDef{
					{Type: "compute.fail", Code: "ERR", Message: "no catch"},
				},
			},
		},
	}

	_, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no catch")
}

func TestEngine_FlowTry_ErrorVariable(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				As:   "result",
				Try: []metadata.CommandDef{
					{Type: "compute.fail", Code: "MY_CODE", Message: "my message"},
				},
				Catch: []metadata.CommandDef{
					{Type: "compute.transform", As: "step1", Value: map[string]string{"err_code": "$.error.code"}},
				},
			},
		},
		Result: map[string]string{
			"err_code": "$.step1.err_code",
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "MY_CODE", result.Result["err_code"])
}

func TestEngine_FlowTry_VarsFromTryVisible(t *testing.T) {
	t.Parallel()

	cache := newTestCELCache()
	computeExec := NewComputeCommandExecutor(NewExpressionResolver(cache))
	eng := NewEngine(cache, nil, computeExec)
	flowExec := NewFlowCommandExecutor(eng, NewExpressionResolver(cache))
	eng.RegisterExecutor(flowExec)

	def := &metadata.ProcedureDefinition{
		Commands: []metadata.CommandDef{
			{
				Type: "flow.try",
				As:   "result",
				Try: []metadata.CommandDef{
					{Type: "compute.transform", As: "step1", Value: map[string]string{"data": "from_try"}},
				},
			},
		},
		Result: map[string]string{
			"data": "$.step1.data",
		},
	}

	result, err := eng.ExecuteDefinition(context.Background(), def, map[string]any{})
	require.NoError(t, err)
	assert.Equal(t, "from_try", result.Result["data"])
}
