package engine

// Limits defines configurable limits for SOQL queries.
// These limits help protect against resource exhaustion and
// ensure queries stay within reasonable bounds.
type Limits struct {
	// MaxSelectFields is the maximum number of fields allowed in SELECT clause.
	// 0 means no limit.
	MaxSelectFields int

	// MaxRecords is the maximum number of records returned by a query.
	// This value is used as default LIMIT if not specified in query.
	// 0 means no limit (not recommended for production).
	MaxRecords int

	// MaxOffset is the maximum allowed OFFSET value.
	// High offsets can cause performance issues.
	// 0 means no limit.
	MaxOffset int

	// MaxLookupDepth is the maximum depth of Child-to-Parent lookups.
	// Example: Contact.Account.Owner.Manager.Name has depth 4.
	// 0 means no limit.
	MaxLookupDepth int

	// MaxSubqueries is the maximum number of Parent-to-Child subqueries
	// allowed in a single SELECT clause.
	// 0 means no limit.
	MaxSubqueries int

	// MaxSubqueryRecords is the maximum number of records returned
	// by each Parent-to-Child subquery per parent record.
	// 0 means no limit.
	MaxSubqueryRecords int

	// MaxQueryLength is the maximum length of the SOQL query string in characters.
	// 0 means no limit.
	MaxQueryLength int

	// MaxGroupByFields is the maximum number of fields in GROUP BY clause.
	// 0 means no limit.
	MaxGroupByFields int

	// MaxOrderByFields is the maximum number of fields in ORDER BY clause.
	// 0 means no limit.
	MaxOrderByFields int

	// DefaultLimit is the default LIMIT value if not specified in query.
	// If 0, no default limit is applied (MaxRecords still enforced).
	DefaultLimit int
}

// DefaultLimits contains the default limits matching Salesforce SOQL.
var DefaultLimits = Limits{
	MaxSelectFields:    0,      // No limit
	MaxRecords:         50000,  // 50,000 records max
	MaxOffset:          2000,   // 2,000 max offset
	MaxLookupDepth:     5,      // 5 levels of parent relationships
	MaxSubqueries:      20,     // 20 subqueries per query
	MaxSubqueryRecords: 200,    // 200 child records per parent
	MaxQueryLength:     100000, // 100,000 characters
	MaxGroupByFields:   0,      // No limit
	MaxOrderByFields:   0,      // No limit
	DefaultLimit:       0,      // No default limit
}

// StrictLimits contains stricter limits suitable for production APIs.
var StrictLimits = Limits{
	MaxSelectFields:    50,    // 50 fields max
	MaxRecords:         1000,  // 1,000 records max
	MaxOffset:          500,   // 500 max offset
	MaxLookupDepth:     3,     // 3 levels max
	MaxSubqueries:      5,     // 5 subqueries max
	MaxSubqueryRecords: 50,    // 50 child records per parent
	MaxQueryLength:     10000, // 10,000 characters
	MaxGroupByFields:   5,     // 5 GROUP BY fields
	MaxOrderByFields:   3,     // 3 ORDER BY fields
	DefaultLimit:       100,   // Default to 100 records
}

// NoLimits disables all limits. Use with caution.
var NoLimits = Limits{}

// CheckSelectFields checks if the number of select fields exceeds the limit.
func (l *Limits) CheckSelectFields(count int) error {
	if l.MaxSelectFields > 0 && count > l.MaxSelectFields {
		return NewLimitError(LimitTypeMaxFields, l.MaxSelectFields, count)
	}
	return nil
}

// CheckRecords checks if the number of records exceeds the limit.
func (l *Limits) CheckRecords(count int) error {
	if l.MaxRecords > 0 && count > l.MaxRecords {
		return NewLimitError(LimitTypeMaxRecords, l.MaxRecords, count)
	}
	return nil
}

// CheckOffset checks if the offset value exceeds the limit.
func (l *Limits) CheckOffset(offset int) error {
	if l.MaxOffset > 0 && offset > l.MaxOffset {
		return NewLimitError(LimitTypeMaxOffset, l.MaxOffset, offset)
	}
	return nil
}

// CheckLookupDepth checks if the lookup depth exceeds the limit.
func (l *Limits) CheckLookupDepth(depth int) error {
	if l.MaxLookupDepth > 0 && depth > l.MaxLookupDepth {
		return NewLimitError(LimitTypeMaxLookupDepth, l.MaxLookupDepth, depth)
	}
	return nil
}

// CheckSubqueries checks if the number of subqueries exceeds the limit.
func (l *Limits) CheckSubqueries(count int) error {
	if l.MaxSubqueries > 0 && count > l.MaxSubqueries {
		return NewLimitError(LimitTypeMaxSubqueries, l.MaxSubqueries, count)
	}
	return nil
}

// CheckSubqueryRecords checks if the subquery records count exceeds the limit.
func (l *Limits) CheckSubqueryRecords(count int) error {
	if l.MaxSubqueryRecords > 0 && count > l.MaxSubqueryRecords {
		return NewLimitError(LimitTypeMaxSubqueryRecords, l.MaxSubqueryRecords, count)
	}
	return nil
}

// CheckQueryLength checks if the query length exceeds the limit.
func (l *Limits) CheckQueryLength(length int) error {
	if l.MaxQueryLength > 0 && length > l.MaxQueryLength {
		return NewLimitError(LimitTypeMaxQueryLength, l.MaxQueryLength, length)
	}
	return nil
}

// EffectiveLimit returns the effective LIMIT value for a query.
// It considers the requested limit, default limit, and max records.
func (l *Limits) EffectiveLimit(requested *int) int {
	var limit int

	if requested != nil && *requested > 0 {
		limit = *requested
	} else if l.DefaultLimit > 0 {
		limit = l.DefaultLimit
	} else if l.MaxRecords > 0 {
		limit = l.MaxRecords
	} else {
		return 0 // No limit
	}

	// Ensure limit doesn't exceed MaxRecords
	if l.MaxRecords > 0 && limit > l.MaxRecords {
		limit = l.MaxRecords
	}

	return limit
}

// EffectiveSubqueryLimit returns the effective LIMIT for a subquery.
func (l *Limits) EffectiveSubqueryLimit(requested *int) int {
	var limit int

	if requested != nil && *requested > 0 {
		limit = *requested
	} else if l.MaxSubqueryRecords > 0 {
		limit = l.MaxSubqueryRecords
	} else {
		return 0 // No limit
	}

	// Ensure limit doesn't exceed MaxSubqueryRecords
	if l.MaxSubqueryRecords > 0 && limit > l.MaxSubqueryRecords {
		limit = l.MaxSubqueryRecords
	}

	return limit
}

// Merge combines two Limits, preferring non-zero values from the override.
func (l *Limits) Merge(override *Limits) *Limits {
	if override == nil {
		return l
	}

	result := *l

	if override.MaxSelectFields > 0 {
		result.MaxSelectFields = override.MaxSelectFields
	}
	if override.MaxRecords > 0 {
		result.MaxRecords = override.MaxRecords
	}
	if override.MaxOffset > 0 {
		result.MaxOffset = override.MaxOffset
	}
	if override.MaxLookupDepth > 0 {
		result.MaxLookupDepth = override.MaxLookupDepth
	}
	if override.MaxSubqueries > 0 {
		result.MaxSubqueries = override.MaxSubqueries
	}
	if override.MaxSubqueryRecords > 0 {
		result.MaxSubqueryRecords = override.MaxSubqueryRecords
	}
	if override.MaxQueryLength > 0 {
		result.MaxQueryLength = override.MaxQueryLength
	}
	if override.MaxGroupByFields > 0 {
		result.MaxGroupByFields = override.MaxGroupByFields
	}
	if override.MaxOrderByFields > 0 {
		result.MaxOrderByFields = override.MaxOrderByFields
	}
	if override.DefaultLimit > 0 {
		result.DefaultLimit = override.DefaultLimit
	}

	return &result
}
