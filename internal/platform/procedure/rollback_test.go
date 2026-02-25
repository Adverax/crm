package procedure

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecuteRollback_LIFOOrder(t *testing.T) {
	t.Parallel()

	var order []string
	execCtx := &ExecutionContext{
		RollbackStack: []RollbackEntry{
			{StepName: "first", Action: func() error { order = append(order, "first"); return nil }},
			{StepName: "second", Action: func() error { order = append(order, "second"); return nil }},
			{StepName: "third", Action: func() error { order = append(order, "third"); return nil }},
		},
	}

	err := ExecuteRollback(execCtx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"third", "second", "first"}, order)
}

func TestExecuteRollback_ContinuesOnError(t *testing.T) {
	t.Parallel()

	var executed []string
	execCtx := &ExecutionContext{
		RollbackStack: []RollbackEntry{
			{StepName: "a", Action: func() error { executed = append(executed, "a"); return nil }},
			{StepName: "b", Action: func() error { executed = append(executed, "b"); return errors.New("b failed") }},
			{StepName: "c", Action: func() error { executed = append(executed, "c"); return nil }},
		},
	}

	err := ExecuteRollback(execCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "b failed")
	assert.Equal(t, []string{"c", "b", "a"}, executed)
}

func TestExecuteRollback_EmptyStack(t *testing.T) {
	t.Parallel()

	execCtx := &ExecutionContext{
		RollbackStack: nil,
	}

	err := ExecuteRollback(execCtx)
	assert.NoError(t, err)
}
