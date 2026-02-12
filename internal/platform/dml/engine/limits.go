package engine

// Limits defines configurable limits for DML operations.
// These limits help protect against resource exhaustion and
// ensure operations stay within reasonable bounds.
type Limits struct {
	// MaxBatchSize is the maximum number of rows in a single INSERT/UPSERT VALUES clause.
	// 0 means no limit.
	MaxBatchSize int

	// MaxFieldsPerRow is the maximum number of fields per row in INSERT/UPDATE.
	// 0 means no limit.
	MaxFieldsPerRow int

	// MaxStatementLength is the maximum length of the DML statement string in characters.
	// 0 means no limit.
	MaxStatementLength int

	// RequireWhereOnDelete if true, DELETE statements must have a WHERE clause.
	RequireWhereOnDelete bool

	// RequireWhereOnUpdate if true, UPDATE statements must have a WHERE clause.
	RequireWhereOnUpdate bool
}

// DefaultLimits contains the default limits for DML operations.
var DefaultLimits = Limits{
	MaxBatchSize:         10000, // 10,000 rows per batch
	MaxFieldsPerRow:      0,     // No limit
	MaxStatementLength:   100000,
	RequireWhereOnDelete: true, // Safety: require WHERE on DELETE
	RequireWhereOnUpdate: false,
}

// StrictLimits contains stricter limits suitable for production APIs.
var StrictLimits = Limits{
	MaxBatchSize:         1000,  // 1,000 rows per batch
	MaxFieldsPerRow:      50,    // 50 fields max per row
	MaxStatementLength:   50000, // 50K characters
	RequireWhereOnDelete: true,
	RequireWhereOnUpdate: true, // Also require WHERE on UPDATE
}

// NoLimits disables all limits. Use with caution.
var NoLimits = Limits{
	RequireWhereOnDelete: false,
	RequireWhereOnUpdate: false,
}

// CheckBatchSize checks if the batch size exceeds the limit.
func (l *Limits) CheckBatchSize(count int) error {
	if l.MaxBatchSize > 0 && count > l.MaxBatchSize {
		return NewLimitError(LimitTypeMaxBatchSize, l.MaxBatchSize, count)
	}
	return nil
}

// CheckFieldsPerRow checks if the number of fields per row exceeds the limit.
func (l *Limits) CheckFieldsPerRow(count int) error {
	if l.MaxFieldsPerRow > 0 && count > l.MaxFieldsPerRow {
		return NewLimitError(LimitTypeMaxFieldsPerRow, l.MaxFieldsPerRow, count)
	}
	return nil
}

// CheckStatementLength checks if the statement length exceeds the limit.
func (l *Limits) CheckStatementLength(length int) error {
	if l.MaxStatementLength > 0 && length > l.MaxStatementLength {
		return NewLimitError(LimitTypeMaxStatementLength, l.MaxStatementLength, length)
	}
	return nil
}

// Merge combines two Limits, preferring non-zero values from the override.
func (l *Limits) Merge(override *Limits) *Limits {
	if override == nil {
		return l
	}

	result := *l

	if override.MaxBatchSize > 0 {
		result.MaxBatchSize = override.MaxBatchSize
	}
	if override.MaxFieldsPerRow > 0 {
		result.MaxFieldsPerRow = override.MaxFieldsPerRow
	}
	if override.MaxStatementLength > 0 {
		result.MaxStatementLength = override.MaxStatementLength
	}
	// Boolean flags are always taken from override
	result.RequireWhereOnDelete = override.RequireWhereOnDelete
	result.RequireWhereOnUpdate = override.RequireWhereOnUpdate

	return &result
}
