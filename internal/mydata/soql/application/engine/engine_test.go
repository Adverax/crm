package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEngineBasic(t *testing.T) {
	metadata := setupTestMetadata()
	engine := NewEngine(WithMetadata(metadata))
	ctx := context.Background()

	t.Run("Prepare simple query", func(t *testing.T) {
		compiled, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Fatalf("Prepare() error = %v", err)
		}

		if compiled.SQL == "" {
			t.Error("expected non-empty SQL")
		}

		if compiled.Shape == nil {
			t.Error("expected non-nil Shape")
		}

		if compiled.Shape.Object != "Account" {
			t.Errorf("Shape.Object = %s, want Account", compiled.Shape.Object)
		}
	})

	t.Run("Prepare complex query", func(t *testing.T) {
		query := `
			SELECT Name, Email, Account.Name
			FROM Contact
			WHERE Email IS NOT NULL
			ORDER BY Name ASC
			LIMIT 100
		`
		compiled, err := engine.Prepare(ctx, query)
		if err != nil {
			t.Fatalf("Prepare() error = %v", err)
		}

		if !strings.Contains(compiled.SQL, "SELECT") {
			t.Error("expected SQL to contain SELECT")
		}

		if !strings.Contains(compiled.SQL, "LEFT JOIN") {
			t.Error("expected SQL to contain LEFT JOIN for Account lookup")
		}
	})

	t.Run("Parse error", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "INVALID QUERY")
		if err == nil {
			t.Error("expected error for invalid query")
		}

		if !IsParseError(err) {
			t.Errorf("expected ParseError, got %T: %v", err, err)
		}
	})

	t.Run("Unknown object", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM UnknownObject")
		if err == nil {
			t.Error("expected error for unknown object")
		}

		if !IsValidationError(err) {
			t.Errorf("expected ValidationError, got %T: %v", err, err)
		}
	})

	t.Run("Unknown field", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT UnknownField FROM Account")
		if err == nil {
			t.Error("expected error for unknown field")
		}

		if !IsValidationError(err) {
			t.Errorf("expected ValidationError, got %T: %v", err, err)
		}
	})
}

func TestEngineWithAccessControl(t *testing.T) {
	metadata := setupTestMetadata()

	// Access controller that only allows Account and User
	access := &ObjectAccessController{
		AllowedObjects: map[string]bool{
			"Account": true,
			"User":    true,
		},
	}

	engine := NewEngine(
		WithMetadata(metadata),
		WithAccessController(access),
	)
	ctx := context.Background()

	t.Run("allowed object", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Errorf("expected no error for allowed object, got %v", err)
		}
	})

	t.Run("denied object", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Contact")
		if err == nil {
			t.Error("expected error for denied object")
		}

		if !IsAccessError(err) {
			t.Errorf("expected AccessError, got %T: %v", err, err)
		}
	})
}

func TestEngineWithLimits(t *testing.T) {
	metadata := setupTestMetadata()

	limits := &Limits{
		MaxRecords:     100,
		MaxOffset:      50,
		MaxLookupDepth: 2,
	}

	engine := NewEngine(
		WithMetadata(metadata),
		WithLimits(limits),
	)
	ctx := context.Background()

	t.Run("within limits", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account LIMIT 50 OFFSET 20")
		if err != nil {
			t.Errorf("expected no error within limits, got %v", err)
		}
	})

	t.Run("exceeds limit", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account LIMIT 200")
		if err == nil {
			t.Error("expected error for exceeding limit")
		}

		if !IsLimitError(err) {
			t.Errorf("expected LimitError, got %T: %v", err, err)
		}
	})

	t.Run("exceeds offset", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account LIMIT 10 OFFSET 100")
		if err == nil {
			t.Error("expected error for exceeding offset")
		}

		if !IsLimitError(err) {
			t.Errorf("expected LimitError, got %T: %v", err, err)
		}
	})

	t.Run("exceeds lookup depth", func(t *testing.T) {
		// Account.Owner.Manager.Name is 3 levels, exceeds limit of 2
		_, err := engine.Prepare(ctx, "SELECT Account.Owner.Manager.Name FROM Contact")
		if err == nil {
			t.Error("expected error for exceeding lookup depth")
		}

		if !IsLimitError(err) {
			t.Errorf("expected LimitError, got %T: %v", err, err)
		}
	})
}

func TestEngineWithDateResolver(t *testing.T) {
	metadata := setupTestMetadata()

	// Fixed time resolver
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)
	dateResolver := &DefaultDateResolver{
		Now:          func() time.Time { return fixedNow },
		Location:     time.UTC,
		WeekStartsOn: time.Monday,
	}

	engine := NewEngine(
		WithMetadata(metadata),
		WithDateResolver(dateResolver),
	)
	ctx := context.Background()

	t.Run("resolve date literal", func(t *testing.T) {
		compiled, err := engine.PrepareAndResolve(ctx, "SELECT Name FROM Account WHERE CreatedDate = TODAY")
		if err != nil {
			t.Fatalf("PrepareAndResolve() error = %v", err)
		}

		// Check that the date param was resolved
		if len(compiled.Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(compiled.Params))
		}

		resolvedDate, ok := compiled.Params[0].(time.Time)
		if !ok {
			t.Fatalf("expected time.Time param, got %T", compiled.Params[0])
		}

		expected := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		if !resolvedDate.Equal(expected) {
			t.Errorf("resolved date = %v, want %v", resolvedDate, expected)
		}
	})

	t.Run("resolve dynamic date literal", func(t *testing.T) {
		compiled, err := engine.PrepareAndResolve(ctx, "SELECT Name FROM Account WHERE CreatedDate >= LAST_N_DAYS:30")
		if err != nil {
			t.Fatalf("PrepareAndResolve() error = %v", err)
		}

		if len(compiled.Params) != 1 {
			t.Fatalf("expected 1 param, got %d", len(compiled.Params))
		}

		resolvedDate, ok := compiled.Params[0].(time.Time)
		if !ok {
			t.Fatalf("expected time.Time param, got %T", compiled.Params[0])
		}

		// 30 days before 2024-03-15 is 2024-02-14
		expected := time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC)
		if !resolvedDate.Equal(expected) {
			t.Errorf("resolved date = %v, want %v", resolvedDate, expected)
		}
	})
}

func TestQueryBuilder(t *testing.T) {
	metadata := setupTestMetadata()
	engine := NewEngine(WithMetadata(metadata))

	t.Run("basic query builder", func(t *testing.T) {
		compiled, err := engine.Query("SELECT Name FROM Account").Prepare()
		if err != nil {
			t.Fatalf("Prepare() error = %v", err)
		}

		if compiled.SQL == "" {
			t.Error("expected non-empty SQL")
		}
	})

	t.Run("query builder with context", func(t *testing.T) {
		ctx := context.Background()
		compiled, err := engine.Query("SELECT Name FROM Account").
			WithContext(ctx).
			Prepare()
		if err != nil {
			t.Fatalf("Prepare() error = %v", err)
		}

		if compiled.SQL == "" {
			t.Error("expected non-empty SQL")
		}
	})
}

func TestStandaloneFunctions(t *testing.T) {
	t.Run("ParseOnly", func(t *testing.T) {
		ast, err := ParseOnly("SELECT Name FROM Account")
		if err != nil {
			t.Fatalf("ParseOnly() error = %v", err)
		}

		if ast.From != "Account" {
			t.Errorf("From = %s, want Account", ast.From)
		}
	})

	t.Run("ParseOnly invalid", func(t *testing.T) {
		_, err := ParseOnly("INVALID")
		if err == nil {
			t.Error("expected error for invalid query")
		}
	})

	t.Run("ValidateOnly", func(t *testing.T) {
		metadata := setupTestMetadata()
		ast, _ := ParseOnly("SELECT Name FROM Account")

		validated, err := ValidateOnly(context.Background(), ast, metadata, nil, nil)
		if err != nil {
			t.Fatalf("ValidateOnly() error = %v", err)
		}

		if validated.RootObject.Name != "Account" {
			t.Errorf("RootObject.Name = %s, want Account", validated.RootObject.Name)
		}
	})

	t.Run("CompileOnly", func(t *testing.T) {
		metadata := setupTestMetadata()
		ast, _ := ParseOnly("SELECT Name FROM Account")
		validated, _ := ValidateOnly(context.Background(), ast, metadata, nil, nil)

		compiled, err := CompileOnly(validated, nil)
		if err != nil {
			t.Fatalf("CompileOnly() error = %v", err)
		}

		if !strings.Contains(compiled.SQL, "SELECT") {
			t.Error("expected SQL to contain SELECT")
		}
	})
}

func TestEngineGetters(t *testing.T) {
	metadata := setupTestMetadata()
	limits := &Limits{MaxRecords: 1000}
	dateResolver := NewDefaultDateResolver()

	engine := NewEngine(
		WithMetadata(metadata),
		WithLimits(limits),
		WithDateResolver(dateResolver),
	)

	t.Run("GetMetadata", func(t *testing.T) {
		if engine.GetMetadata() != metadata {
			t.Error("GetMetadata() returned wrong value")
		}
	})

	t.Run("GetLimits", func(t *testing.T) {
		if engine.GetLimits() != limits {
			t.Error("GetLimits() returned wrong value")
		}
	})

	t.Run("GetDateResolver", func(t *testing.T) {
		if engine.GetDateResolver() != dateResolver {
			t.Error("GetDateResolver() returned wrong value")
		}
	})
}

func TestEngineSetters(t *testing.T) {
	metadata := setupTestMetadata()
	engine := NewEngine(WithMetadata(metadata))
	ctx := context.Background()

	t.Run("SetMetadata", func(t *testing.T) {
		newMetadata := setupTestMetadata()
		engine.SetMetadata(newMetadata)

		if engine.GetMetadata() != newMetadata {
			t.Error("SetMetadata() did not update metadata")
		}

		// Verify it still works
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Errorf("Prepare() error after SetMetadata = %v", err)
		}
	})

	t.Run("SetLimits", func(t *testing.T) {
		newLimits := &Limits{MaxRecords: 500}
		engine.SetLimits(newLimits)

		if engine.GetLimits() != newLimits {
			t.Error("SetLimits() did not update limits")
		}

		// Verify limit is enforced
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account LIMIT 600")
		if err == nil {
			t.Error("expected error after SetLimits")
		}
	})
}

func TestMustParse(t *testing.T) {
	metadata := setupTestMetadata()
	engine := NewEngine(WithMetadata(metadata))

	t.Run("valid query", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustParse panicked unexpectedly: %v", r)
			}
		}()

		ast := engine.MustParse("SELECT Name FROM Account")
		if ast.From != "Account" {
			t.Errorf("From = %s, want Account", ast.From)
		}
	})

	t.Run("invalid query panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected MustParse to panic on invalid query")
			}
		}()

		engine.MustParse("INVALID")
	})
}

func TestQueryLengthLimit(t *testing.T) {
	metadata := setupTestMetadata()
	limits := &Limits{MaxQueryLength: 50}
	engine := NewEngine(
		WithMetadata(metadata),
		WithLimits(limits),
	)
	ctx := context.Background()

	t.Run("within limit", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Errorf("expected no error for short query, got %v", err)
		}
	})

	t.Run("exceeds limit", func(t *testing.T) {
		longQuery := "SELECT Name, Industry, AnnualRevenue, CreatedDate, OwnerId FROM Account"
		_, err := engine.Prepare(ctx, longQuery)
		if err == nil {
			t.Error("expected error for long query")
		}

		if !IsLimitError(err) {
			t.Errorf("expected LimitError, got %T: %v", err, err)
		}
	})
}

func TestErrorUnwrapping(t *testing.T) {
	metadata := setupTestMetadata()
	engine := NewEngine(WithMetadata(metadata))
	ctx := context.Background()

	t.Run("ParseError unwrapping", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "INVALID QUERY")
		if err == nil {
			t.Fatal("expected error")
		}

		var parseErr *ParseError
		if !errors.As(err, &parseErr) {
			t.Errorf("expected to unwrap ParseError, got %T", err)
		}
	})

	t.Run("ValidationError unwrapping", func(t *testing.T) {
		_, err := engine.Prepare(ctx, "SELECT Unknown FROM Account")
		if err == nil {
			t.Fatal("expected error")
		}

		var validErr *ValidationError
		if !errors.As(err, &validErr) {
			t.Errorf("expected to unwrap ValidationError, got %T", err)
		}
	})
}

func TestQueryCache(t *testing.T) {
	metadata := setupTestMetadata()
	ctx := context.Background()

	t.Run("cache disabled by default", func(t *testing.T) {
		engine := NewEngine(WithMetadata(metadata))
		if engine.IsCacheEnabled() {
			t.Error("expected cache to be disabled by default (no cache injected)")
		}
	})

	t.Run("cache enabled with test cache", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)
		if !engine.IsCacheEnabled() {
			t.Error("expected cache to be enabled")
		}
	})

	t.Run("cache disabled with nil option", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(nil),
		)
		if engine.IsCacheEnabled() {
			t.Error("expected cache to be disabled")
		}
	})

	t.Run("cache hit returns same result", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)

		// First call - cache miss
		compiled1, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Fatalf("first Prepare() error = %v", err)
		}

		// Second call - cache hit
		compiled2, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Fatalf("second Prepare() error = %v", err)
		}

		// Should return the same pointer (cached)
		if compiled1 != compiled2 {
			t.Error("expected same compiled query from cache")
		}
	})

	t.Run("different queries different results", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)

		compiled1, err := engine.Prepare(ctx, "SELECT Name FROM Account")
		if err != nil {
			t.Fatalf("first Prepare() error = %v", err)
		}

		compiled2, err := engine.Prepare(ctx, "SELECT Name FROM Contact")
		if err != nil {
			t.Fatalf("second Prepare() error = %v", err)
		}

		// Different queries should have different results
		if compiled1 == compiled2 {
			t.Error("expected different compiled queries for different SOQL")
		}
	})

	t.Run("cache stats", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)

		// Make some queries
		if _, err := engine.Prepare(ctx, "SELECT Name FROM Account"); err != nil {
			t.Fatalf("first Prepare() error = %v", err)
		}
		if _, err := engine.Prepare(ctx, "SELECT Name FROM Account"); err != nil {
			t.Fatalf("second Prepare() error = %v", err)
		}
		if _, err := engine.Prepare(ctx, "SELECT Name FROM Contact"); err != nil {
			t.Fatalf("third Prepare() error = %v", err)
		}

		stats := engine.QueryCacheStats()
		if stats == nil {
			t.Fatal("expected non-nil cache stats")
		}

		// Check that we have 2 entries (Account and Contact queries)
		if stats.Size != 2 {
			t.Errorf("expected 2 entries, got %d", stats.Size)
		}
	})

	t.Run("invalidate query", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)
		query := "SELECT Name FROM Account"

		// Cache the query
		compiled1, _ := engine.Prepare(ctx, query)

		// Invalidate it
		engine.InvalidateQuery(ctx, query)

		// Should compile again (different pointer)
		compiled2, _ := engine.Prepare(ctx, query)

		if compiled1 == compiled2 {
			t.Error("expected different compiled query after invalidation")
		}
	})

	t.Run("clear cache", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)

		// Cache some queries
		compiled1, _ := engine.Prepare(ctx, "SELECT Name FROM Account")
		_, _ = engine.Prepare(ctx, "SELECT Name FROM Contact")

		// Clear cache
		engine.ClearQueryCache(ctx)

		// Should compile again
		compiled2, _ := engine.Prepare(ctx, "SELECT Name FROM Account")

		if compiled1 == compiled2 {
			t.Error("expected different compiled query after clear")
		}
	})

	t.Run("SetMetadata clears cache", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(newTestCache[string, *CompiledQuery]()),
		)

		// Cache a query
		compiled1, _ := engine.Prepare(ctx, "SELECT Name FROM Account")

		// Change metadata (should clear cache)
		engine.SetMetadata(setupTestMetadata())

		// Should compile again
		compiled2, _ := engine.Prepare(ctx, "SELECT Name FROM Account")

		if compiled1 == compiled2 {
			t.Error("expected different compiled query after SetMetadata")
		}
	})

	t.Run("cache disabled does not cache", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(nil),
		)

		// First call
		compiled1, _ := engine.Prepare(ctx, "SELECT Name FROM Account")

		// Second call - should be different (not cached)
		compiled2, _ := engine.Prepare(ctx, "SELECT Name FROM Account")

		if compiled1 == compiled2 {
			t.Error("expected different compiled queries when cache disabled")
		}
	})

	t.Run("cache stats nil when disabled", func(t *testing.T) {
		engine := NewEngine(
			WithMetadata(metadata),
			WithQueryCache(nil),
		)

		stats := engine.QueryCacheStats()
		if stats != nil {
			t.Error("expected nil stats when cache disabled")
		}
	})
}
