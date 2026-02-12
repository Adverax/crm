package engine

import (
	"testing"
)

func TestDefaultLimits(t *testing.T) {
	l := DefaultLimits

	if l.MaxRecords != 50000 {
		t.Errorf("DefaultLimits.MaxRecords = %d, want 50000", l.MaxRecords)
	}
	if l.MaxOffset != 2000 {
		t.Errorf("DefaultLimits.MaxOffset = %d, want 2000", l.MaxOffset)
	}
	if l.MaxLookupDepth != 5 {
		t.Errorf("DefaultLimits.MaxLookupDepth = %d, want 5", l.MaxLookupDepth)
	}
	if l.MaxSubqueries != 20 {
		t.Errorf("DefaultLimits.MaxSubqueries = %d, want 20", l.MaxSubqueries)
	}
	if l.MaxSubqueryRecords != 200 {
		t.Errorf("DefaultLimits.MaxSubqueryRecords = %d, want 200", l.MaxSubqueryRecords)
	}
	if l.MaxQueryLength != 100000 {
		t.Errorf("DefaultLimits.MaxQueryLength = %d, want 100000", l.MaxQueryLength)
	}
}

func TestStrictLimits(t *testing.T) {
	l := StrictLimits

	if l.MaxSelectFields != 50 {
		t.Errorf("StrictLimits.MaxSelectFields = %d, want 50", l.MaxSelectFields)
	}
	if l.MaxRecords != 1000 {
		t.Errorf("StrictLimits.MaxRecords = %d, want 1000", l.MaxRecords)
	}
	if l.MaxOffset != 500 {
		t.Errorf("StrictLimits.MaxOffset = %d, want 500", l.MaxOffset)
	}
	if l.DefaultLimit != 100 {
		t.Errorf("StrictLimits.DefaultLimit = %d, want 100", l.DefaultLimit)
	}
}

func TestNoLimits(t *testing.T) {
	l := NoLimits

	if l.MaxRecords != 0 {
		t.Errorf("NoLimits.MaxRecords = %d, want 0", l.MaxRecords)
	}
	if l.MaxOffset != 0 {
		t.Errorf("NoLimits.MaxOffset = %d, want 0", l.MaxOffset)
	}
}

func TestCheckSelectFields(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		count     int
		wantError bool
	}{
		{"within limit", 50, 25, false},
		{"at limit", 50, 50, false},
		{"exceeds limit", 50, 51, true},
		{"no limit (0)", 0, 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxSelectFields: tt.limit}
			err := l.CheckSelectFields(tt.count)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckSelectFields(%d) error = %v, wantError %v", tt.count, err, tt.wantError)
			}
			if err != nil && !IsLimitError(err) {
				t.Errorf("CheckSelectFields(%d) should return LimitError", tt.count)
			}
		})
	}
}

func TestCheckRecords(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		count     int
		wantError bool
	}{
		{"within limit", 1000, 500, false},
		{"at limit", 1000, 1000, false},
		{"exceeds limit", 1000, 1001, true},
		{"no limit (0)", 0, 100000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxRecords: tt.limit}
			err := l.CheckRecords(tt.count)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckRecords(%d) error = %v, wantError %v", tt.count, err, tt.wantError)
			}
		})
	}
}

func TestCheckOffset(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		wantError bool
	}{
		{"within limit", 2000, 1000, false},
		{"at limit", 2000, 2000, false},
		{"exceeds limit", 2000, 2001, true},
		{"no limit (0)", 0, 100000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxOffset: tt.limit}
			err := l.CheckOffset(tt.offset)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckOffset(%d) error = %v, wantError %v", tt.offset, err, tt.wantError)
			}
		})
	}
}

func TestCheckLookupDepth(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		depth     int
		wantError bool
	}{
		{"within limit", 5, 3, false},
		{"at limit", 5, 5, false},
		{"exceeds limit", 5, 6, true},
		{"no limit (0)", 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxLookupDepth: tt.limit}
			err := l.CheckLookupDepth(tt.depth)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckLookupDepth(%d) error = %v, wantError %v", tt.depth, err, tt.wantError)
			}
		})
	}
}

func TestCheckSubqueries(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		count     int
		wantError bool
	}{
		{"within limit", 20, 10, false},
		{"at limit", 20, 20, false},
		{"exceeds limit", 20, 21, true},
		{"no limit (0)", 0, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxSubqueries: tt.limit}
			err := l.CheckSubqueries(tt.count)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckSubqueries(%d) error = %v, wantError %v", tt.count, err, tt.wantError)
			}
		})
	}
}

func TestCheckSubqueryRecords(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		count     int
		wantError bool
	}{
		{"within limit", 200, 100, false},
		{"at limit", 200, 200, false},
		{"exceeds limit", 200, 201, true},
		{"no limit (0)", 0, 1000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxSubqueryRecords: tt.limit}
			err := l.CheckSubqueryRecords(tt.count)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckSubqueryRecords(%d) error = %v, wantError %v", tt.count, err, tt.wantError)
			}
		})
	}
}

func TestCheckQueryLength(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		length    int
		wantError bool
	}{
		{"within limit", 10000, 5000, false},
		{"at limit", 10000, 10000, false},
		{"exceeds limit", 10000, 10001, true},
		{"no limit (0)", 0, 1000000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Limits{MaxQueryLength: tt.limit}
			err := l.CheckQueryLength(tt.length)
			if (err != nil) != tt.wantError {
				t.Errorf("CheckQueryLength(%d) error = %v, wantError %v", tt.length, err, tt.wantError)
			}
		})
	}
}

func TestEffectiveLimit(t *testing.T) {
	tests := []struct {
		name          string
		limits        Limits
		requested     *int
		wantEffective int
	}{
		{
			name:          "requested limit used",
			limits:        Limits{MaxRecords: 1000, DefaultLimit: 100},
			requested:     intPtrLimits(50),
			wantEffective: 50,
		},
		{
			name:          "requested exceeds max, capped",
			limits:        Limits{MaxRecords: 1000, DefaultLimit: 100},
			requested:     intPtrLimits(5000),
			wantEffective: 1000,
		},
		{
			name:          "nil requested, default used",
			limits:        Limits{MaxRecords: 1000, DefaultLimit: 100},
			requested:     nil,
			wantEffective: 100,
		},
		{
			name:          "zero requested, default used",
			limits:        Limits{MaxRecords: 1000, DefaultLimit: 100},
			requested:     intPtrLimits(0),
			wantEffective: 100,
		},
		{
			name:          "no default, max used",
			limits:        Limits{MaxRecords: 1000, DefaultLimit: 0},
			requested:     nil,
			wantEffective: 1000,
		},
		{
			name:          "no limits, returns 0",
			limits:        Limits{MaxRecords: 0, DefaultLimit: 0},
			requested:     nil,
			wantEffective: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.limits.EffectiveLimit(tt.requested)
			if got != tt.wantEffective {
				t.Errorf("EffectiveLimit() = %d, want %d", got, tt.wantEffective)
			}
		})
	}
}

func TestEffectiveSubqueryLimit(t *testing.T) {
	tests := []struct {
		name          string
		limits        Limits
		requested     *int
		wantEffective int
	}{
		{
			name:          "requested limit used",
			limits:        Limits{MaxSubqueryRecords: 200},
			requested:     intPtrLimits(50),
			wantEffective: 50,
		},
		{
			name:          "requested exceeds max, capped",
			limits:        Limits{MaxSubqueryRecords: 200},
			requested:     intPtrLimits(500),
			wantEffective: 200,
		},
		{
			name:          "nil requested, max used",
			limits:        Limits{MaxSubqueryRecords: 200},
			requested:     nil,
			wantEffective: 200,
		},
		{
			name:          "no limit, returns 0",
			limits:        Limits{MaxSubqueryRecords: 0},
			requested:     nil,
			wantEffective: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.limits.EffectiveSubqueryLimit(tt.requested)
			if got != tt.wantEffective {
				t.Errorf("EffectiveSubqueryLimit() = %d, want %d", got, tt.wantEffective)
			}
		})
	}
}

func TestLimitsMerge(t *testing.T) {
	base := &Limits{
		MaxSelectFields:    100,
		MaxRecords:         50000,
		MaxOffset:          2000,
		MaxLookupDepth:     5,
		MaxSubqueries:      20,
		MaxSubqueryRecords: 200,
		MaxQueryLength:     100000,
		MaxGroupByFields:   10,
		MaxOrderByFields:   5,
		DefaultLimit:       1000,
	}

	t.Run("nil override returns base", func(t *testing.T) {
		result := base.Merge(nil)
		if result.MaxRecords != 50000 {
			t.Errorf("Merge(nil).MaxRecords = %d, want 50000", result.MaxRecords)
		}
	})

	t.Run("override values replace base", func(t *testing.T) {
		override := &Limits{
			MaxRecords:   1000,
			DefaultLimit: 100,
		}
		result := base.Merge(override)

		if result.MaxRecords != 1000 {
			t.Errorf("Merge().MaxRecords = %d, want 1000", result.MaxRecords)
		}
		if result.DefaultLimit != 100 {
			t.Errorf("Merge().DefaultLimit = %d, want 100", result.DefaultLimit)
		}
		// Base values should be preserved for non-overridden fields
		if result.MaxOffset != 2000 {
			t.Errorf("Merge().MaxOffset = %d, want 2000", result.MaxOffset)
		}
	})

	t.Run("zero values in override don't replace", func(t *testing.T) {
		override := &Limits{
			MaxRecords: 1000,
			MaxOffset:  0, // Should not replace
		}
		result := base.Merge(override)

		if result.MaxOffset != 2000 {
			t.Errorf("Merge().MaxOffset = %d, want 2000 (zero should not replace)", result.MaxOffset)
		}
	})

	t.Run("all fields override", func(t *testing.T) {
		override := &Limits{
			MaxSelectFields:    10,
			MaxRecords:         500,
			MaxOffset:          100,
			MaxLookupDepth:     2,
			MaxSubqueries:      5,
			MaxSubqueryRecords: 50,
			MaxQueryLength:     5000,
			MaxGroupByFields:   3,
			MaxOrderByFields:   2,
			DefaultLimit:       50,
		}
		result := base.Merge(override)

		if result.MaxSelectFields != 10 {
			t.Errorf("Merge().MaxSelectFields = %d, want 10", result.MaxSelectFields)
		}
		if result.MaxRecords != 500 {
			t.Errorf("Merge().MaxRecords = %d, want 500", result.MaxRecords)
		}
		if result.MaxOffset != 100 {
			t.Errorf("Merge().MaxOffset = %d, want 100", result.MaxOffset)
		}
		if result.MaxLookupDepth != 2 {
			t.Errorf("Merge().MaxLookupDepth = %d, want 2", result.MaxLookupDepth)
		}
		if result.MaxSubqueries != 5 {
			t.Errorf("Merge().MaxSubqueries = %d, want 5", result.MaxSubqueries)
		}
		if result.MaxSubqueryRecords != 50 {
			t.Errorf("Merge().MaxSubqueryRecords = %d, want 50", result.MaxSubqueryRecords)
		}
		if result.MaxQueryLength != 5000 {
			t.Errorf("Merge().MaxQueryLength = %d, want 5000", result.MaxQueryLength)
		}
		if result.MaxGroupByFields != 3 {
			t.Errorf("Merge().MaxGroupByFields = %d, want 3", result.MaxGroupByFields)
		}
		if result.MaxOrderByFields != 2 {
			t.Errorf("Merge().MaxOrderByFields = %d, want 2", result.MaxOrderByFields)
		}
		if result.DefaultLimit != 50 {
			t.Errorf("Merge().DefaultLimit = %d, want 50", result.DefaultLimit)
		}
	})
}

func TestLimitErrorDetails(t *testing.T) {
	l := &Limits{MaxRecords: 1000}
	err := l.CheckRecords(5000)

	if err == nil {
		t.Fatal("expected error")
	}

	limitErr, ok := err.(*LimitError)
	if !ok {
		t.Fatalf("expected *LimitError, got %T", err)
	}

	if limitErr.LimitType != LimitTypeMaxRecords {
		t.Errorf("LimitType = %v, want LimitTypeMaxRecords", limitErr.LimitType)
	}
	if limitErr.Limit != 1000 {
		t.Errorf("Limit = %d, want 1000", limitErr.Limit)
	}
	if limitErr.Actual != 5000 {
		t.Errorf("Actual = %d, want 5000", limitErr.Actual)
	}
}

func intPtrLimits(i int) *int {
	return &i
}
