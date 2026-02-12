package engine

import (
	"context"
	"fmt"
)

// Engine is the main entry point for DML statement processing.
// It coordinates parsing, validation, and compilation of DML statements.
type Engine struct {
	// Core dependencies
	metadata MetadataProvider
	access   WriteAccessController
	limits   *Limits

	// Internal components
	validator *Validator
	compiler  *Compiler

	// Executor for running compiled statements (optional)
	executor Executor
}

// Option configures an Engine.
type Option func(*Engine)

// WithMetadata sets the metadata provider for the engine.
func WithMetadata(m MetadataProvider) Option {
	return func(e *Engine) {
		e.metadata = m
	}
}

// WithWriteAccessController sets the write access controller for the engine.
func WithWriteAccessController(a WriteAccessController) Option {
	return func(e *Engine) {
		e.access = a
	}
}

// WithLimits sets the operation limits for the engine.
func WithLimits(l *Limits) Option {
	return func(e *Engine) {
		e.limits = l
	}
}

// WithExecutor sets the executor for running compiled statements.
func WithExecutor(ex Executor) Option {
	return func(e *Engine) {
		e.executor = ex
	}
}

// NewEngineFromDependencies creates a new DML Engine from a Dependencies container.
// This is the preferred way to create an Engine with full dependency injection.
func NewEngineFromDependencies(deps *Dependencies, opts ...Option) *Engine {
	e := &Engine{
		limits: &DefaultLimits,
	}

	// Apply dependencies
	if deps != nil {
		e.metadata = deps.MetadataProvider
		e.access = deps.WriteAccessController
		e.executor = deps.Executor
	}

	// Apply options (can override dependencies)
	for _, opt := range opts {
		opt(e)
	}

	// Initialize defaults for missing dependencies
	if e.access == nil {
		e.access = &NoopWriteAccessController{}
	}

	// Create validator and compiler
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.compiler = NewCompiler(e.limits)

	return e
}

// NewEngine creates a new DML Engine with the given options.
// For full dependency injection, use NewEngineFromDependencies instead.
func NewEngine(opts ...Option) *Engine {
	return NewEngineFromDependencies(nil, opts...)
}

// Parse parses a DML statement string into an AST.
// Returns a ParseError if the statement has syntax errors.
func (e *Engine) Parse(statement string) (*DMLStatement, error) {
	// Check statement length limit
	if e.limits != nil && e.limits.MaxStatementLength > 0 && len(statement) > e.limits.MaxStatementLength {
		return nil, NewLimitError(LimitTypeMaxStatementLength, e.limits.MaxStatementLength, len(statement))
	}

	ast, err := Parse(statement)
	if err != nil {
		return nil, NewParseErrorFromParticiple(err)
	}

	return ast, nil
}

// Validate validates a parsed DML AST against metadata and access rules.
// Returns a ValidationError or AccessError if validation fails.
func (e *Engine) Validate(ctx context.Context, ast *DMLStatement) (*ValidatedDML, error) {
	if e.metadata == nil {
		return nil, fmt.Errorf("metadata provider is required for validation")
	}
	return e.validator.Validate(ctx, ast)
}

// Compile compiles a validated DML statement into SQL.
func (e *Engine) Compile(validated *ValidatedDML) (*CompiledDML, error) {
	return e.compiler.Compile(validated)
}

// Prepare parses, validates, and compiles a DML statement in a single call.
// This is the main method for preparing a statement for execution.
func (e *Engine) Prepare(ctx context.Context, statement string) (*CompiledDML, error) {
	// Parse
	ast, err := e.Parse(statement)
	if err != nil {
		return nil, err
	}

	// Validate
	validated, err := e.Validate(ctx, ast)
	if err != nil {
		return nil, err
	}

	// Compile
	compiled, err := e.Compile(validated)
	if err != nil {
		return nil, err
	}

	return compiled, nil
}

// Execute prepares and executes a DML statement in a single call.
// Returns an error if no executor is configured.
func (e *Engine) Execute(ctx context.Context, statement string) (*Result, error) {
	if e.executor == nil {
		return nil, fmt.Errorf("no executor configured")
	}

	compiled, err := e.Prepare(ctx, statement)
	if err != nil {
		return nil, err
	}

	return e.executor.Execute(ctx, compiled)
}

// ExecuteCompiled executes a pre-compiled DML statement.
// Returns an error if no executor is configured.
func (e *Engine) ExecuteCompiled(ctx context.Context, compiled *CompiledDML) (*Result, error) {
	if e.executor == nil {
		return nil, fmt.Errorf("no executor configured")
	}

	return e.executor.Execute(ctx, compiled)
}

// MustParse parses a DML statement and panics on error.
// Useful for statements that are known at compile time.
func (e *Engine) MustParse(statement string) *DMLStatement {
	ast, err := e.Parse(statement)
	if err != nil {
		panic(err)
	}
	return ast
}

// GetMetadata returns the engine's metadata provider.
func (e *Engine) GetMetadata() MetadataProvider {
	return e.metadata
}

// GetLimits returns the engine's limits configuration.
func (e *Engine) GetLimits() *Limits {
	return e.limits
}

// SetMetadata updates the engine's metadata provider.
// This recreates the validator with the new metadata.
func (e *Engine) SetMetadata(m MetadataProvider) {
	e.metadata = m
	e.validator = NewValidator(e.metadata, e.access, e.limits)
}

// SetWriteAccessController updates the engine's write access controller.
// This recreates the validator with the new access controller.
func (e *Engine) SetWriteAccessController(a WriteAccessController) {
	e.access = a
	e.validator = NewValidator(e.metadata, e.access, e.limits)
}

// SetLimits updates the engine's limits configuration.
// This recreates both the validator and compiler with new limits.
func (e *Engine) SetLimits(l *Limits) {
	e.limits = l
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.compiler = NewCompiler(e.limits)
}

// SetExecutor updates the engine's executor.
func (e *Engine) SetExecutor(ex Executor) {
	e.executor = ex
}

// StatementBuilder provides a fluent API for building and executing statements.
type StatementBuilder struct {
	engine    *Engine
	statement string
	ctx       context.Context
}

// Statement starts building a statement.
func (e *Engine) Statement(statement string) *StatementBuilder {
	return &StatementBuilder{
		engine:    e,
		statement: statement,
		ctx:       context.Background(),
	}
}

// WithContext sets the context for the statement builder.
func (b *StatementBuilder) WithContext(ctx context.Context) *StatementBuilder {
	b.ctx = ctx
	return b
}

// Prepare parses, validates, and compiles the statement.
func (b *StatementBuilder) Prepare() (*CompiledDML, error) {
	return b.engine.Prepare(b.ctx, b.statement)
}

// Execute prepares and executes the statement.
func (b *StatementBuilder) Execute() (*Result, error) {
	return b.engine.Execute(b.ctx, b.statement)
}

// ParseOnly parses a DML statement without validation.
// Useful for syntax checking without metadata.
func ParseOnly(statement string) (*DMLStatement, error) {
	return Parse(statement)
}

// ValidateOnly validates an AST with the given metadata provider.
// Useful for validation without full engine setup.
func ValidateOnly(ctx context.Context, ast *DMLStatement, metadata MetadataProvider, access WriteAccessController, limits *Limits) (*ValidatedDML, error) {
	validator := NewValidator(metadata, access, limits)
	return validator.Validate(ctx, ast)
}

// CompileOnly compiles a validated statement with the given limits.
// Useful for compilation without full engine setup.
func CompileOnly(validated *ValidatedDML, limits *Limits) (*CompiledDML, error) {
	compiler := NewCompiler(limits)
	return compiler.Compile(validated)
}
