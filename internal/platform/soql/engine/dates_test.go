package engine

import (
	"context"
	"testing"
	"time"
)

func TestResolveStaticDateLiterals(t *testing.T) {
	// Fixed time for testing: Monday, 2024-03-15 12:00:00 UTC
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:          func() time.Time { return fixedNow },
		Location:     time.UTC,
		WeekStartsOn: time.Monday,
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   StaticDateLiteral
		wantStart string // Expected start date (YYYY-MM-DD)
		wantEnd   string // Expected end date (YYYY-MM-DD)
	}{
		{
			name:      "TODAY",
			literal:   DateToday,
			wantStart: "2024-03-15",
			wantEnd:   "2024-03-15",
		},
		{
			name:      "YESTERDAY",
			literal:   DateYesterday,
			wantStart: "2024-03-14",
			wantEnd:   "2024-03-14",
		},
		{
			name:      "TOMORROW",
			literal:   DateTomorrow,
			wantStart: "2024-03-16",
			wantEnd:   "2024-03-16",
		},
		{
			name:      "THIS_WEEK",
			literal:   DateThisWeek,
			wantStart: "2024-03-11", // Monday
			wantEnd:   "2024-03-17", // Sunday
		},
		{
			name:      "LAST_WEEK",
			literal:   DateLastWeek,
			wantStart: "2024-03-04", // Monday of previous week
			wantEnd:   "2024-03-10", // Sunday of previous week
		},
		{
			name:      "NEXT_WEEK",
			literal:   DateNextWeek,
			wantStart: "2024-03-18", // Monday of next week
			wantEnd:   "2024-03-24", // Sunday of next week
		},
		{
			name:      "THIS_MONTH",
			literal:   DateThisMonth,
			wantStart: "2024-03-01",
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_MONTH",
			literal:   DateLastMonth,
			wantStart: "2024-02-01",
			wantEnd:   "2024-02-29", // 2024 is leap year
		},
		{
			name:      "NEXT_MONTH",
			literal:   DateNextMonth,
			wantStart: "2024-04-01",
			wantEnd:   "2024-04-30",
		},
		{
			name:      "THIS_QUARTER",
			literal:   DateThisQuarter,
			wantStart: "2024-01-01", // Q1 starts in January
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_QUARTER",
			literal:   DateLastQuarter,
			wantStart: "2023-10-01", // Q4 2023
			wantEnd:   "2023-12-31",
		},
		{
			name:      "THIS_YEAR",
			literal:   DateThisYear,
			wantStart: "2024-01-01",
			wantEnd:   "2024-12-31",
		},
		{
			name:      "LAST_YEAR",
			literal:   DateLastYear,
			wantStart: "2023-01-01",
			wantEnd:   "2023-12-31",
		},
		{
			name:      "NEXT_YEAR",
			literal:   DateNextYear,
			wantStart: "2025-01-01",
			wantEnd:   "2025-12-31",
		},
		{
			name:      "LAST_90_DAYS",
			literal:   DateLast90Days,
			wantStart: "2023-12-16", // 90 days before 2024-03-15
			wantEnd:   "2024-03-15",
		},
		{
			name:      "NEXT_90_DAYS",
			literal:   DateNext90Days,
			wantStart: "2024-03-15",
			wantEnd:   "2024-06-13", // 90 days after 2024-03-15
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveStaticRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveStaticRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestResolveDynamicDateLiterals(t *testing.T) {
	// Fixed time for testing: Monday, 2024-03-15 12:00:00 UTC
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:          func() time.Time { return fixedNow },
		Location:     time.UTC,
		WeekStartsOn: time.Monday,
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   *DynamicDateLiteral
		wantStart string
		wantEnd   string
	}{
		{
			name: "LAST_N_DAYS:7",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNDays,
				N:    7,
			},
			wantStart: "2024-03-08",
			wantEnd:   "2024-03-15",
		},
		{
			name: "LAST_N_DAYS:30",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNDays,
				N:    30,
			},
			wantStart: "2024-02-14",
			wantEnd:   "2024-03-15",
		},
		{
			name: "NEXT_N_DAYS:7",
			literal: &DynamicDateLiteral{
				Type: DynamicNextNDays,
				N:    7,
			},
			wantStart: "2024-03-15",
			wantEnd:   "2024-03-22",
		},
		{
			name: "LAST_N_MONTHS:3",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNMonths,
				N:    3,
			},
			wantStart: "2023-12-01",
			wantEnd:   "2024-02-29",
		},
		{
			name: "NEXT_N_MONTHS:2",
			literal: &DynamicDateLiteral{
				Type: DynamicNextNMonths,
				N:    2,
			},
			wantStart: "2024-04-01",
			wantEnd:   "2024-05-31",
		},
		{
			name: "LAST_N_YEARS:1",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNYears,
				N:    1,
			},
			wantStart: "2023-01-01",
			wantEnd:   "2023-12-31",
		},
		{
			name: "NEXT_N_YEARS:1",
			literal: &DynamicDateLiteral{
				Type: DynamicNextNYears,
				N:    1,
			},
			wantStart: "2025-01-01",
			wantEnd:   "2025-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveDynamicRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveDynamicRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestResolveDateParams(t *testing.T) {
	// Fixed time for testing
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:          func() time.Time { return fixedNow },
		Location:     time.UTC,
		WeekStartsOn: time.Monday,
	}
	ctx := context.Background()

	// Create a compiled query with date params
	today := DateToday
	query := &CompiledQuery{
		SQL: "SELECT * FROM accounts WHERE created_at = $1",
		Params: []any{
			nil, // placeholder for date
		},
		DateParams: []*DateParam{
			{
				ParamIndex: 1,
				Static:     &today,
			},
		},
	}

	// Resolve params
	err := ResolveDateParams(ctx, query, resolver)
	if err != nil {
		t.Fatalf("ResolveDateParams() error = %v", err)
	}

	// Check the resolved date
	resolvedDate, ok := query.Params[0].(time.Time)
	if !ok {
		t.Fatalf("Params[0] is not time.Time, got %T", query.Params[0])
	}

	expected := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	if !resolvedDate.Equal(expected) {
		t.Errorf("resolved date = %v, want %v", resolvedDate, expected)
	}
}

func TestResolveDateParamsWithDynamic(t *testing.T) {
	// Fixed time for testing
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:          func() time.Time { return fixedNow },
		Location:     time.UTC,
		WeekStartsOn: time.Monday,
	}
	ctx := context.Background()

	// Create a compiled query with dynamic date params
	query := &CompiledQuery{
		SQL: "SELECT * FROM accounts WHERE created_at >= $1",
		Params: []any{
			nil, // placeholder for date
		},
		DateParams: []*DateParam{
			{
				ParamIndex: 1,
				Dynamic: &DynamicDateLiteral{
					Type: DynamicLastNDays,
					N:    30,
				},
			},
		},
	}

	// Resolve params
	err := ResolveDateParams(ctx, query, resolver)
	if err != nil {
		t.Fatalf("ResolveDateParams() error = %v", err)
	}

	// Check the resolved date
	resolvedDate, ok := query.Params[0].(time.Time)
	if !ok {
		t.Fatalf("Params[0] is not time.Time, got %T", query.Params[0])
	}

	// Should be 30 days before 2024-03-15
	expected := time.Date(2024, 2, 14, 0, 0, 0, 0, time.UTC)
	if !resolvedDate.Equal(expected) {
		t.Errorf("resolved date = %v, want %v", resolvedDate, expected)
	}
}

func TestDefaultDateResolver(t *testing.T) {
	// Test that default resolver works with real time
	resolver := NewDefaultDateResolver()
	ctx := context.Background()

	start, err := resolver.ResolveStatic(ctx, DateToday)
	if err != nil {
		t.Fatalf("ResolveStatic(TODAY) error = %v", err)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	startTruncated := start.Truncate(24 * time.Hour)

	if !startTruncated.Equal(today) {
		t.Errorf("TODAY resolved to %v, expected around %v", start, today)
	}
}

func TestResolverWithTimezone(t *testing.T) {
	// Test timezone handling
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skip("America/New_York timezone not available")
	}

	resolver := NewDateResolverWithLocation(loc)
	ctx := context.Background()

	start, err := resolver.ResolveStatic(ctx, DateToday)
	if err != nil {
		t.Fatalf("ResolveStatic(TODAY) error = %v", err)
	}

	// Verify the timezone is correct
	if start.Location().String() != "America/New_York" {
		t.Errorf("timezone = %s, want America/New_York", start.Location().String())
	}
}

func TestFiscalPeriods_CalendarYearDefault(t *testing.T) {
	// Test fiscal periods when fiscal year equals calendar year (default)
	// Fixed time: 2024-03-15 (Q1, mid-March)
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:                  func() time.Time { return fixedNow },
		Location:             time.UTC,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.January, // Fiscal year = calendar year
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   StaticDateLiteral
		wantStart string
		wantEnd   string
	}{
		{
			name:      "THIS_FISCAL_QUARTER (Q1)",
			literal:   DateThisFiscalQuarter,
			wantStart: "2024-01-01",
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_FISCAL_QUARTER (Q4 prev year)",
			literal:   DateLastFiscalQuarter,
			wantStart: "2023-10-01",
			wantEnd:   "2023-12-31",
		},
		{
			name:      "NEXT_FISCAL_QUARTER (Q2)",
			literal:   DateNextFiscalQuarter,
			wantStart: "2024-04-01",
			wantEnd:   "2024-06-30",
		},
		{
			name:      "THIS_FISCAL_YEAR",
			literal:   DateThisFiscalYear,
			wantStart: "2024-01-01",
			wantEnd:   "2024-12-31",
		},
		{
			name:      "LAST_FISCAL_YEAR",
			literal:   DateLastFiscalYear,
			wantStart: "2023-01-01",
			wantEnd:   "2023-12-31",
		},
		{
			name:      "NEXT_FISCAL_YEAR",
			literal:   DateNextFiscalYear,
			wantStart: "2025-01-01",
			wantEnd:   "2025-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveStaticRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveStaticRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestFiscalPeriods_AprilFiscalYear(t *testing.T) {
	// Test fiscal periods when fiscal year starts in April
	// Fixed time: 2024-03-15 (FY Q4 in April FY, since March is before April)
	// Fiscal year 2023 runs from April 2023 to March 2024
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:                  func() time.Time { return fixedNow },
		Location:             time.UTC,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.April, // Fiscal year starts in April
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   StaticDateLiteral
		wantStart string
		wantEnd   string
	}{
		{
			name:      "THIS_FISCAL_QUARTER (FQ4 = Jan-Mar)",
			literal:   DateThisFiscalQuarter,
			wantStart: "2024-01-01",
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_FISCAL_QUARTER (FQ3 = Oct-Dec)",
			literal:   DateLastFiscalQuarter,
			wantStart: "2023-10-01",
			wantEnd:   "2023-12-31",
		},
		{
			name:      "NEXT_FISCAL_QUARTER (FQ1 next FY = Apr-Jun)",
			literal:   DateNextFiscalQuarter,
			wantStart: "2024-04-01",
			wantEnd:   "2024-06-30",
		},
		{
			name:      "THIS_FISCAL_YEAR (Apr 2023 - Mar 2024)",
			literal:   DateThisFiscalYear,
			wantStart: "2023-04-01",
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_FISCAL_YEAR (Apr 2022 - Mar 2023)",
			literal:   DateLastFiscalYear,
			wantStart: "2022-04-01",
			wantEnd:   "2023-03-31",
		},
		{
			name:      "NEXT_FISCAL_YEAR (Apr 2024 - Mar 2025)",
			literal:   DateNextFiscalYear,
			wantStart: "2024-04-01",
			wantEnd:   "2025-03-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveStaticRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveStaticRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestFiscalPeriods_OctoberFiscalYear(t *testing.T) {
	// Test fiscal periods when fiscal year starts in October (US Government FY)
	// Fixed time: 2024-03-15 (FY Q2 in October FY)
	// Fiscal year 2024 runs from October 2023 to September 2024
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:                  func() time.Time { return fixedNow },
		Location:             time.UTC,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.October, // US Government FY starts in October
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   StaticDateLiteral
		wantStart string
		wantEnd   string
	}{
		{
			name:      "THIS_FISCAL_QUARTER (FQ2 = Jan-Mar)",
			literal:   DateThisFiscalQuarter,
			wantStart: "2024-01-01",
			wantEnd:   "2024-03-31",
		},
		{
			name:      "LAST_FISCAL_QUARTER (FQ1 = Oct-Dec)",
			literal:   DateLastFiscalQuarter,
			wantStart: "2023-10-01",
			wantEnd:   "2023-12-31",
		},
		{
			name:      "NEXT_FISCAL_QUARTER (FQ3 = Apr-Jun)",
			literal:   DateNextFiscalQuarter,
			wantStart: "2024-04-01",
			wantEnd:   "2024-06-30",
		},
		{
			name:      "THIS_FISCAL_YEAR (Oct 2023 - Sep 2024)",
			literal:   DateThisFiscalYear,
			wantStart: "2023-10-01",
			wantEnd:   "2024-09-30",
		},
		{
			name:      "LAST_FISCAL_YEAR (Oct 2022 - Sep 2023)",
			literal:   DateLastFiscalYear,
			wantStart: "2022-10-01",
			wantEnd:   "2023-09-30",
		},
		{
			name:      "NEXT_FISCAL_YEAR (Oct 2024 - Sep 2025)",
			literal:   DateNextFiscalYear,
			wantStart: "2024-10-01",
			wantEnd:   "2025-09-30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveStaticRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveStaticRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestDynamicFiscalPeriods(t *testing.T) {
	// Test dynamic fiscal periods with April fiscal year
	fixedNow := time.Date(2024, 3, 15, 12, 0, 0, 0, time.UTC)

	resolver := &DefaultDateResolver{
		Now:                  func() time.Time { return fixedNow },
		Location:             time.UTC,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.April,
	}
	ctx := context.Background()

	tests := []struct {
		name      string
		literal   *DynamicDateLiteral
		wantStart string
		wantEnd   string
	}{
		{
			name: "LAST_N_FISCAL_QUARTERS:2",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNFiscalQuarters,
				N:    2,
			},
			wantStart: "2023-07-01", // 2 quarters before FQ4 (Jan-Mar) = FQ2 (Jul-Sep)
			wantEnd:   "2023-12-31", // End of FQ3 (Oct-Dec)
		},
		{
			name: "NEXT_N_FISCAL_QUARTERS:2",
			literal: &DynamicDateLiteral{
				Type: DynamicNextNFiscalQuarters,
				N:    2,
			},
			wantStart: "2024-04-01", // Next FQ after FQ4 = FQ1 of next FY
			wantEnd:   "2024-09-30", // End of FQ2
		},
		{
			name: "LAST_N_FISCAL_YEARS:2",
			literal: &DynamicDateLiteral{
				Type: DynamicLastNFiscalYears,
				N:    2,
			},
			wantStart: "2021-04-01", // 2 FYs before FY2023 (Apr 2023 - Mar 2024) = FY2021 (Apr 2021)
			wantEnd:   "2023-03-31", // End of FY2022
		},
		{
			name: "NEXT_N_FISCAL_YEARS:2",
			literal: &DynamicDateLiteral{
				Type: DynamicNextNFiscalYears,
				N:    2,
			},
			wantStart: "2024-04-01", // Next FY starts Apr 2024
			wantEnd:   "2026-03-31", // End of FY2025
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, err := resolver.ResolveDynamicRange(ctx, tt.literal)
			if err != nil {
				t.Fatalf("ResolveDynamicRange() error = %v", err)
			}

			gotStart := start.Format("2006-01-02")
			gotEnd := end.Format("2006-01-02")

			if gotStart != tt.wantStart {
				t.Errorf("start = %s, want %s", gotStart, tt.wantStart)
			}

			if gotEnd != tt.wantEnd {
				t.Errorf("end = %s, want %s", gotEnd, tt.wantEnd)
			}
		})
	}
}

func TestNewDateResolverWithFiscalYear(t *testing.T) {
	resolver := NewDateResolverWithFiscalYear(time.UTC, time.April)
	if resolver.FiscalYearStartMonth != time.April {
		t.Errorf("FiscalYearStartMonth = %v, want %v", resolver.FiscalYearStartMonth, time.April)
	}
	if resolver.WeekStartsOn != time.Monday {
		t.Errorf("WeekStartsOn = %v, want %v", resolver.WeekStartsOn, time.Monday)
	}
}

func TestFiscalQuarterNumber(t *testing.T) {
	resolver := &DefaultDateResolver{
		FiscalYearStartMonth: time.April,
	}

	tests := []struct {
		month time.Month
		want  int
	}{
		{time.April, 1},
		{time.May, 1},
		{time.June, 1},
		{time.July, 2},
		{time.August, 2},
		{time.September, 2},
		{time.October, 3},
		{time.November, 3},
		{time.December, 3},
		{time.January, 4},
		{time.February, 4},
		{time.March, 4},
	}

	for _, tt := range tests {
		t.Run(tt.month.String(), func(t *testing.T) {
			// Use a fixed date in that month
			date := time.Date(2024, tt.month, 15, 12, 0, 0, 0, time.UTC)
			got := resolver.fiscalQuarterNumber(date)
			if got != tt.want {
				t.Errorf("fiscalQuarterNumber(%v) = %d, want %d", tt.month, got, tt.want)
			}
		})
	}
}
