package cel

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

// StandardEnv creates a CEL environment with all standard variables:
// record (current values), old (previous values for UPDATE), user, now.
func StandardEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("record", cel.DynType),
		cel.Variable("old", cel.DynType),
		cel.Variable("user", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		ext.Strings(),
	)
}

// DefaultEnv creates a CEL environment for default expressions:
// record (current values), user, now.
func DefaultEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("record", cel.DynType),
		cel.Variable("user", cel.DynType),
		cel.Variable("now", cel.TimestampType),
		ext.Strings(),
	)
}
