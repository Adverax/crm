package metadata

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const (
	maxFunctionCount = 200
	maxNestingDepth  = 3
	maxBodySize      = 4096
	maxParamsCount   = 10
)

var validFunctionName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

var validReturnTypes = map[string]bool{
	"string": true, "number": true, "boolean": true,
	"list": true, "map": true, "any": true,
}

var validParamTypes = map[string]bool{
	"string": true, "number": true, "boolean": true,
	"list": true, "map": true, "any": true,
}

// FunctionService provides business logic for custom functions.
type FunctionService interface {
	Create(ctx context.Context, input CreateFunctionInput) (*Function, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Function, error)
	ListAll(ctx context.Context) ([]Function, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateFunctionInput) (*Function, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// OnFunctionsChanged is a callback invoked after functions are modified.
type OnFunctionsChanged func(ctx context.Context) error

type functionService struct {
	pool     *pgxpool.Pool
	repo     FunctionRepository
	cache    *MetadataCache
	onChange OnFunctionsChanged
}

// NewFunctionService creates a new FunctionService.
func NewFunctionService(
	pool *pgxpool.Pool,
	repo FunctionRepository,
	cache *MetadataCache,
	onChange OnFunctionsChanged,
) FunctionService {
	return &functionService{
		pool:     pool,
		repo:     repo,
		cache:    cache,
		onChange: onChange,
	}
}

func (s *functionService) Create(ctx context.Context, input CreateFunctionInput) (*Function, error) {
	if err := validateFunctionInput(input.Name, input.Body, input.ReturnType, input.Params); err != nil {
		return nil, fmt.Errorf("functionService.Create: %w", err)
	}

	// Check limit
	count, err := s.repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("functionService.Create: %w", err)
	}
	if count >= maxFunctionCount {
		return nil, fmt.Errorf("functionService.Create: %w",
			apperror.BadRequest(fmt.Sprintf("max function limit reached (%d)", maxFunctionCount)))
	}

	// Check name uniqueness
	existing, err := s.repo.GetByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("functionService.Create: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("functionService.Create: %w",
			apperror.Conflict(fmt.Sprintf("function %q already exists", input.Name)))
	}

	// Apply defaults
	if input.ReturnType == "" {
		input.ReturnType = "any"
	}
	if input.Params == nil {
		input.Params = []FunctionParam{}
	}

	// Check cycles and nesting with proposed function
	allFunctions := s.cache.GetFunctions()
	proposed := Function{Name: input.Name, Body: input.Body}
	allWithProposed := append(allFunctions, proposed)

	if err := DetectCycles(allWithProposed); err != nil {
		return nil, fmt.Errorf("functionService.Create: %w",
			apperror.BadRequest("cyclic dependency: "+err.Error()))
	}
	if err := DetectNestingDepth(allWithProposed, maxNestingDepth); err != nil {
		return nil, fmt.Errorf("functionService.Create: %w",
			apperror.BadRequest(err.Error()))
	}

	fn, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("functionService.Create: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("functionService.Create: %w", err)
	}

	return fn, nil
}

func (s *functionService) GetByID(ctx context.Context, id uuid.UUID) (*Function, error) {
	fn, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("functionService.GetByID: %w", err)
	}
	if fn == nil {
		return nil, fmt.Errorf("functionService.GetByID: %w",
			apperror.NotFound("function", id.String()))
	}
	return fn, nil
}

func (s *functionService) ListAll(ctx context.Context) ([]Function, error) {
	functions, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("functionService.ListAll: %w", err)
	}
	return functions, nil
}

func (s *functionService) Update(ctx context.Context, id uuid.UUID, input UpdateFunctionInput) (*Function, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("functionService.Update: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("functionService.Update: %w",
			apperror.NotFound("function", id.String()))
	}

	if err := validateFunctionInput(existing.Name, input.Body, input.ReturnType, input.Params); err != nil {
		return nil, fmt.Errorf("functionService.Update: %w", err)
	}

	if input.ReturnType == "" {
		input.ReturnType = "any"
	}
	if input.Params == nil {
		input.Params = []FunctionParam{}
	}

	// Check cycles and nesting with updated function
	allFunctions := s.cache.GetFunctions()
	var allUpdated []Function
	for _, fn := range allFunctions {
		if fn.ID == id {
			allUpdated = append(allUpdated, Function{Name: fn.Name, Body: input.Body})
		} else {
			allUpdated = append(allUpdated, fn)
		}
	}

	if err := DetectCycles(allUpdated); err != nil {
		return nil, fmt.Errorf("functionService.Update: %w",
			apperror.BadRequest("cyclic dependency: "+err.Error()))
	}
	if err := DetectNestingDepth(allUpdated, maxNestingDepth); err != nil {
		return nil, fmt.Errorf("functionService.Update: %w",
			apperror.BadRequest(err.Error()))
	}

	fn, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, fmt.Errorf("functionService.Update: %w", err)
	}
	if fn == nil {
		return nil, fmt.Errorf("functionService.Update: %w",
			apperror.NotFound("function", id.String()))
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return nil, fmt.Errorf("functionService.Update: %w", err)
	}

	return fn, nil
}

func (s *functionService) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("functionService.Delete: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("functionService.Delete: %w",
			apperror.NotFound("function", id.String()))
	}

	// Check usages (requires DB connection)
	if s.pool != nil {
		usages, err := FindUsages(ctx, s.pool, existing.Name)
		if err != nil {
			return fmt.Errorf("functionService.Delete: %w", err)
		}
		if len(usages) > 0 {
			return fmt.Errorf("functionService.Delete: %w",
				apperror.Conflict(fmt.Sprintf("function %q is used in %d place(s)", existing.Name, len(usages))))
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("functionService.Delete: %w", err)
	}

	if err := s.reloadAndNotify(ctx); err != nil {
		return fmt.Errorf("functionService.Delete: %w", err)
	}

	return nil
}

func (s *functionService) reloadAndNotify(ctx context.Context) error {
	if err := s.cache.LoadFunctions(ctx); err != nil {
		return fmt.Errorf("cache reload: %w", err)
	}
	if s.onChange != nil {
		if err := s.onChange(ctx); err != nil {
			return fmt.Errorf("onChange callback: %w", err)
		}
	}
	return nil
}

func validateFunctionInput(name, body, returnType string, params []FunctionParam) error {
	if !validFunctionName.MatchString(name) {
		return apperror.BadRequest("name must match ^[a-z][a-z0-9_]*$")
	}
	if len(name) > 100 {
		return apperror.BadRequest("name must be at most 100 characters")
	}
	if body == "" {
		return apperror.BadRequest("body is required")
	}
	if len(body) > maxBodySize {
		return apperror.BadRequest(fmt.Sprintf("body must be at most %d characters", maxBodySize))
	}
	if returnType != "" && !validReturnTypes[returnType] {
		return apperror.BadRequest("invalid return_type: " + returnType)
	}
	if len(params) > maxParamsCount {
		return apperror.BadRequest(fmt.Sprintf("at most %d parameters allowed", maxParamsCount))
	}

	seen := make(map[string]bool, len(params))
	for _, p := range params {
		if p.Name == "" {
			return apperror.BadRequest("parameter name is required")
		}
		if !validFunctionName.MatchString(p.Name) {
			return apperror.BadRequest("parameter name must match ^[a-z][a-z0-9_]*$: " + p.Name)
		}
		if p.Type != "" && !validParamTypes[p.Type] {
			return apperror.BadRequest("invalid parameter type: " + p.Type)
		}
		if seen[p.Name] {
			return apperror.BadRequest("duplicate parameter name: " + p.Name)
		}
		seen[p.Name] = true
	}

	return nil
}
