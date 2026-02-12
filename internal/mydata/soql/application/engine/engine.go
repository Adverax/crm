package engine

import (
	"context"
	"fmt"
	"time"
)

// DefaultQueryCacheTTL is the default time-to-live for cached compiled queries.
const DefaultQueryCacheTTL = 5 * time.Minute

// Engine is the main entry point for SOQL query processing.
// It coordinates parsing, validation, and compilation of SOQL queries.
type Engine struct {
	// Core dependencies
	metadata     MetadataProvider
	access       AccessController
	dateResolver DateResolver
	limits       *Limits

	// Internal components
	validator *Validator
	compiler  *Compiler

	// Query cache for compiled queries (optional, injected)
	queryCache Cache[string, *CompiledQuery]
}

// Option configures an Engine.
type Option func(*Engine)

// WithMetadata sets the metadata provider for the engine.
func WithMetadata(m MetadataProvider) Option {
	return func(e *Engine) {
		e.metadata = m
	}
}

// WithAccessController sets the access controller for the engine.
func WithAccessController(a AccessController) Option {
	return func(e *Engine) {
		e.access = a
	}
}

// WithDateResolver sets the date resolver for the engine.
func WithDateResolver(d DateResolver) Option {
	return func(e *Engine) {
		e.dateResolver = d
	}
}

// WithLimits sets the query limits for the engine.
func WithLimits(l *Limits) Option {
	return func(e *Engine) {
		e.limits = l
	}
}

// WithQueryCache sets a pre-configured cache instance for compiled queries.
// If nil, query caching is disabled.
func WithQueryCache(c Cache[string, *CompiledQuery]) Option {
	return func(e *Engine) {
		e.queryCache = c
	}
}

// NewEngineFromDependencies creates a new SOQL Engine from a Dependencies container.
// This is the preferred way to create an Engine with full dependency injection.
func NewEngineFromDependencies(deps *Dependencies, opts ...Option) *Engine {
	e := &Engine{
		limits: &DefaultLimits,
	}

	// Apply dependencies
	if deps != nil {
		e.metadata = deps.MetadataProvider
		e.access = deps.AccessController
		e.queryCache = deps.QueryCache
	}

	// Apply options (can override dependencies)
	for _, opt := range opts {
		opt(e)
	}

	// Initialize defaults for missing dependencies
	if e.access == nil {
		e.access = &NoopAccessController{}
	}
	if e.dateResolver == nil {
		e.dateResolver = NewDefaultDateResolver()
	}

	// Create validator and compiler
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.compiler = NewCompiler(e.limits)

	return e
}

// NewEngine creates a new SOQL Engine with the given options.
// For full dependency injection, use NewEngineFromDependencies instead.
func NewEngine(opts ...Option) *Engine {
	return NewEngineFromDependencies(nil, opts...)
}

// Parse parses a SOQL query string into an AST.
// Returns a ParseError if the query has syntax errors.
func (e *Engine) Parse(query string) (*Grammar, error) {
	// Check query length limit
	if e.limits != nil && e.limits.MaxQueryLength > 0 && len(query) > e.limits.MaxQueryLength {
		return nil, NewLimitError(LimitTypeMaxQueryLength, e.limits.MaxQueryLength, len(query))
	}

	ast, err := Parse(query)
	if err != nil {
		return nil, NewParseErrorFromParticiple(err)
	}

	return ast, nil
}

// Validate validates a parsed SOQL AST against metadata and access rules.
// Returns a ValidationError or AccessError if validation fails.
func (e *Engine) Validate(ctx context.Context, ast *Grammar) (*ValidatedQuery, error) {
	if e.metadata == nil {
		return nil, fmt.Errorf("metadata provider is required for validation")
	}
	return e.validator.Validate(ctx, ast)
}

// Compile compiles a validated SOQL query into SQL.
func (e *Engine) Compile(validated *ValidatedQuery) (*CompiledQuery, error) {
	return e.compiler.Compile(validated)
}

// Prepare parses, validates, and compiles a SOQL query in a single call.
// This is the main method for preparing a query for execution.
// Results are cached if a query cache was provided.
func (e *Engine) Prepare(ctx context.Context, query string) (*CompiledQuery, error) {
	// Try to get from cache
	if e.queryCache != nil {
		if cached, found := e.queryCache.Get(ctx, query); found {
			return cached, nil
		}
	}

	// Parse
	ast, err := e.Parse(query)
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

	// Cache the result
	if e.queryCache != nil {
		_ = e.queryCache.Set(ctx, query, compiled)
	}

	return compiled, nil
}

// PrepareAndResolve parses, validates, compiles, and resolves date parameters.
// Returns a CompiledQuery with all date parameters resolved to actual values.
func (e *Engine) PrepareAndResolve(ctx context.Context, query string) (*CompiledQuery, error) {
	compiled, err := e.Prepare(ctx, query)
	if err != nil {
		return nil, err
	}

	// Resolve date parameters
	if len(compiled.DateParams) > 0 {
		if err := ResolveDateParams(ctx, compiled, e.dateResolver); err != nil {
			return nil, fmt.Errorf("failed to resolve date parameters: %w", err)
		}
	}

	return compiled, nil
}

// MustParse parses a SOQL query and panics on error.
// Useful for queries that are known at compile time.
func (e *Engine) MustParse(query string) *Grammar {
	ast, err := e.Parse(query)
	if err != nil {
		panic(err)
	}
	return ast
}

// QueryBuilder provides a fluent API for building queries.
type QueryBuilder struct {
	engine *Engine
	query  string
	ctx    context.Context
}

// Query starts building a query.
func (e *Engine) Query(query string) *QueryBuilder {
	return &QueryBuilder{
		engine: e,
		query:  query,
		ctx:    context.Background(),
	}
}

// WithContext sets the context for the query builder.
func (b *QueryBuilder) WithContext(ctx context.Context) *QueryBuilder {
	b.ctx = ctx
	return b
}

// Prepare parses, validates, and compiles the query.
func (b *QueryBuilder) Prepare() (*CompiledQuery, error) {
	return b.engine.Prepare(b.ctx, b.query)
}

// PrepareAndResolve prepares and resolves date parameters.
func (b *QueryBuilder) PrepareAndResolve() (*CompiledQuery, error) {
	return b.engine.PrepareAndResolve(b.ctx, b.query)
}

// GetMetadata returns the engine's metadata provider.
func (e *Engine) GetMetadata() MetadataProvider {
	return e.metadata
}

// GetLimits returns the engine's limits configuration.
func (e *Engine) GetLimits() *Limits {
	return e.limits
}

// GetDateResolver returns the engine's date resolver.
func (e *Engine) GetDateResolver() DateResolver {
	return e.dateResolver
}

// SetMetadata updates the engine's metadata provider.
// This recreates the validator with the new metadata and clears the query cache.
func (e *Engine) SetMetadata(m MetadataProvider) {
	e.metadata = m
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.ClearQueryCache(context.Background())
}

// SetAccessController updates the engine's access controller.
// This recreates the validator with the new access controller and clears the query cache.
func (e *Engine) SetAccessController(a AccessController) {
	e.access = a
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.ClearQueryCache(context.Background())
}

// SetLimits updates the engine's limits configuration.
// This recreates both the validator and compiler with new limits and clears the query cache.
func (e *Engine) SetLimits(l *Limits) {
	e.limits = l
	e.validator = NewValidator(e.metadata, e.access, e.limits)
	e.compiler = NewCompiler(e.limits)
	e.ClearQueryCache(context.Background())
}

// ClearQueryCache clears all cached compiled queries.
func (e *Engine) ClearQueryCache(ctx context.Context) {
	if e.queryCache != nil {
		_ = e.queryCache.Clear(ctx)
	}
}

// InvalidateQuery removes a specific query from the cache.
func (e *Engine) InvalidateQuery(ctx context.Context, query string) {
	if e.queryCache != nil {
		_ = e.queryCache.Delete(ctx, query)
	}
}

// InvalidateQueriesByObject removes all cached queries that depend on the given object.
// This enables targeted cache invalidation when metadata for a specific object changes.
func (e *Engine) InvalidateQueriesByObject(ctx context.Context, objectApiName string) {
	if e.queryCache == nil {
		return
	}
	_ = e.queryCache.Remove(func(q *CompiledQuery) bool {
		for _, dep := range q.Dependencies {
			if dep == objectApiName {
				return true
			}
		}
		return false
	})
}

// QueryCacheStats returns statistics about the query cache.
func (e *Engine) QueryCacheStats() *CacheStats {
	if e.queryCache != nil {
		return e.queryCache.GetStats()
	}
	return nil
}

// StartQueryCacheGC starts the garbage collector for the query cache.
// Only works if the cache implements CacheWithGC interface.
func (e *Engine) StartQueryCacheGC(ctx context.Context, interval time.Duration) {
	if e.queryCache != nil {
		if gc, ok := e.queryCache.(CacheWithGC[string, *CompiledQuery]); ok {
			gc.StartGarbageCollector(ctx, interval)
		}
	}
}

// IsCacheEnabled returns whether query caching is enabled.
func (e *Engine) IsCacheEnabled() bool {
	return e.queryCache != nil
}

// ParseOnly parses a SOQL query without validation.
// Useful for syntax checking without metadata.
func ParseOnly(query string) (*Grammar, error) {
	return Parse(query)
}

// ValidateOnly validates an AST with the given metadata provider.
// Useful for validation without full engine setup.
func ValidateOnly(ctx context.Context, ast *Grammar, metadata MetadataProvider, access AccessController, limits *Limits) (*ValidatedQuery, error) {
	validator := NewValidator(metadata, access, limits)
	return validator.Validate(ctx, ast)
}

// CompileOnly compiles a validated query with the given limits.
// Useful for compilation without full engine setup.
func CompileOnly(validated *ValidatedQuery, limits *Limits) (*CompiledQuery, error) {
	compiler := NewCompiler(limits)
	return compiler.Compile(validated)
}
