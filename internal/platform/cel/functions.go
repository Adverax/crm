package cel

import (
	"fmt"
	"sync"

	gocel "github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

// FunctionDef describes a custom function for registration in a CEL environment.
type FunctionDef struct {
	Name       string
	Params     []ParamDef
	ReturnType string
	Body       string
}

// ParamDef describes a single parameter of a custom function.
type ParamDef struct {
	Name string
	Type string
}

// compiledFunction holds a precompiled CEL program for a custom function body.
type compiledFunction struct {
	program gocel.Program
	params  []ParamDef
}

// FunctionRegistry compiles and caches custom function programs for use in CEL environments.
type FunctionRegistry struct {
	mu        sync.RWMutex
	compiled  map[string]*compiledFunction
	functions []FunctionDef
}

// NewFunctionRegistry creates a new FunctionRegistry and precompiles all function bodies.
func NewFunctionRegistry(functions []FunctionDef) (*FunctionRegistry, error) {
	r := &FunctionRegistry{
		compiled:  make(map[string]*compiledFunction, len(functions)),
		functions: functions,
	}

	if err := r.compile(); err != nil {
		return nil, err
	}
	return r, nil
}

// compile precompiles all function bodies using a minimal CEL environment.
func (r *FunctionRegistry) compile() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, fn := range r.functions {
		env, err := functionBodyEnv(fn.Params)
		if err != nil {
			return fmt.Errorf("functionRegistry.compile(%s): env: %w", fn.Name, err)
		}

		ast, issues := env.Compile(fn.Body)
		if issues != nil && issues.Err() != nil {
			return fmt.Errorf("functionRegistry.compile(%s): %w", fn.Name, issues.Err())
		}

		prog, err := env.Program(ast)
		if err != nil {
			return fmt.Errorf("functionRegistry.compile(%s): program: %w", fn.Name, err)
		}

		r.compiled[fn.Name] = &compiledFunction{
			program: prog,
			params:  fn.Params,
		}
	}
	return nil
}

// evaluate evaluates a named custom function with the given arguments.
func (r *FunctionRegistry) evaluate(name string, args ...ref.Val) ref.Val {
	r.mu.RLock()
	cf, ok := r.compiled[name]
	r.mu.RUnlock()
	if !ok {
		return types.NewErr("unknown function: fn.%s", name)
	}

	if len(args) != len(cf.params) {
		return types.NewErr("fn.%s: expected %d args, got %d", name, len(cf.params), len(args))
	}

	vars := make(map[string]any, len(cf.params))
	for i, p := range cf.params {
		vars[p.Name] = args[i].Value()
	}

	out, _, err := cf.program.Eval(vars)
	if err != nil {
		return types.NewErr("fn.%s: eval: %s", name, err)
	}
	return out
}

// EnvOptions returns gocel.EnvOption entries that register all fn.* functions.
func (r *FunctionRegistry) EnvOptions() []gocel.EnvOption {
	opts := make([]gocel.EnvOption, 0, len(r.functions))
	for _, fn := range r.functions {
		fn := fn // capture
		argTypes := make([]*gocel.Type, len(fn.Params))
		for i := range fn.Params {
			argTypes[i] = gocel.DynType
		}

		overloadID := "fn_" + fn.Name
		opts = append(opts, gocel.Function("fn."+fn.Name,
			gocel.Overload(overloadID, argTypes, gocel.DynType,
				gocel.FunctionBinding(func(args ...ref.Val) ref.Val {
					return r.evaluate(fn.Name, args...)
				}),
			),
		))
	}
	return opts
}

// functionBodyEnv creates a CEL environment for compiling a function body.
// It declares only the function parameters as variables (no record/old/user/now).
func functionBodyEnv(params []ParamDef) (*gocel.Env, error) {
	opts := make([]gocel.EnvOption, 0, len(params)+1)
	for _, p := range params {
		opts = append(opts, gocel.Variable(p.Name, gocel.DynType))
	}
	opts = append(opts, ext.Strings())
	return gocel.NewEnv(opts...)
}

// StandardEnvWithFunctions creates a CEL environment with standard variables + fn.*.
func StandardEnvWithFunctions(registry *FunctionRegistry) (*gocel.Env, error) {
	opts := []gocel.EnvOption{
		gocel.Variable("record", gocel.DynType),
		gocel.Variable("old", gocel.DynType),
		gocel.Variable("user", gocel.DynType),
		gocel.Variable("now", gocel.TimestampType),
		ext.Strings(),
	}
	opts = append(opts, registry.EnvOptions()...)
	return gocel.NewEnv(opts...)
}

// DefaultEnvWithFunctions creates a CEL environment for defaults + fn.*.
func DefaultEnvWithFunctions(registry *FunctionRegistry) (*gocel.Env, error) {
	opts := []gocel.EnvOption{
		gocel.Variable("record", gocel.DynType),
		gocel.Variable("user", gocel.DynType),
		gocel.Variable("now", gocel.TimestampType),
		ext.Strings(),
	}
	opts = append(opts, registry.EnvOptions()...)
	return gocel.NewEnv(opts...)
}

// FunctionBodyEnv creates a CEL environment for validating a function body.
// Parameters are declared as variables; fn.* functions from the registry are also available.
func FunctionBodyEnv(params []ParamDef, registry *FunctionRegistry) (*gocel.Env, error) {
	fnCount := 0
	if registry != nil {
		fnCount = len(registry.functions)
	}
	opts := make([]gocel.EnvOption, 0, len(params)+1+fnCount)
	for _, p := range params {
		opts = append(opts, gocel.Variable(p.Name, gocel.DynType))
	}
	opts = append(opts, ext.Strings())
	if registry != nil {
		opts = append(opts, registry.EnvOptions()...)
	}
	return gocel.NewEnv(opts...)
}
