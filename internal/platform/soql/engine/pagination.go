package engine

// DefaultTieBreaker is the default tie-breaker field for keyset pagination.
// Uses id as the unique identifier for stable pagination (ADR-0001).
const DefaultTieBreaker = "id"

// PaginationConfig holds configuration for keyset pagination.
type PaginationConfig struct {
	// Secret is the secret provider for cursor signing.
	Secret SecretProvider

	// TieBreaker is the field name used as tie-breaker (default: "id").
	TieBreaker string
}

// DefaultPaginationConfig returns default pagination configuration.
// Note: In production, use a proper secret provider.
func DefaultPaginationConfig(secret SecretProvider) *PaginationConfig {
	return &PaginationConfig{
		Secret:     secret,
		TieBreaker: DefaultTieBreaker,
	}
}

// PaginationInfo contains pagination metadata for a compiled query.
type PaginationInfo struct {
	// SortKeys contains the normalized sort keys including tie-breaker.
	// These are SQL column names, not SOQL field names.
	SortKeys SortKeys

	// SortKeySOQL contains the SOQL field names for sort keys.
	// Used for mapping between SOQL and SQL names.
	SortKeySOQL []string

	// TieBreaker is the tie-breaker column name.
	TieBreaker string

	// PageSize is the requested page size.
	PageSize int

	// HasOrderBy indicates if the query has an explicit ORDER BY clause.
	HasOrderBy bool

	// Object is the root object API name (for FID generation).
	Object string
}

// KeysetField maps a SOQL field to its SQL column for pagination.
type KeysetField struct {
	// SOQLName is the SOQL field name (e.g., "CreatedAt").
	SOQLName string

	// SQLColumn is the SQL column name (e.g., "created_at").
	SQLColumn string

	// TableAlias is the table alias in the query (e.g., "t0").
	TableAlias string

	// Direction is the sort direction ("asc" or "desc").
	Direction string
}

// FullColumn returns the fully qualified column name (alias.column).
func (f *KeysetField) FullColumn() string {
	if f.TableAlias == "" {
		return f.SQLColumn
	}
	return f.TableAlias + "." + f.SQLColumn
}
