package engine

import (
	"context"
	"fmt"
	"time"
)

// DateResolver resolves date literals to actual time values.
type DateResolver interface {
	// ResolveStatic resolves a static date literal (TODAY, THIS_WEEK, etc.)
	// Returns the date (or date range start) for the literal.
	ResolveStatic(ctx context.Context, literal StaticDateLiteral) (time.Time, error)

	// ResolveDynamic resolves a dynamic date literal (LAST_N_DAYS:n, etc.)
	// Returns the date (or date range start) for the literal.
	ResolveDynamic(ctx context.Context, literal *DynamicDateLiteral) (time.Time, error)

	// ResolveStaticRange resolves a static date literal to a date range.
	// Returns (start, end) time for literals that represent ranges.
	// For point-in-time literals (TODAY, YESTERDAY), start and end are the same day.
	ResolveStaticRange(ctx context.Context, literal StaticDateLiteral) (start, end time.Time, err error)

	// ResolveDynamicRange resolves a dynamic date literal to a date range.
	ResolveDynamicRange(ctx context.Context, literal *DynamicDateLiteral) (start, end time.Time, err error)
}

// DefaultDateResolver resolves date literals based on the current time.
// Uses a configurable "now" function for testing purposes.
type DefaultDateResolver struct {
	// Now returns the current time. If nil, time.Now() is used.
	Now func() time.Time

	// Location is the timezone for date calculations. If nil, UTC is used.
	Location *time.Location

	// WeekStartsOn specifies which day the week starts on (0=Sunday, 1=Monday).
	// Default is Monday (1) to match ISO 8601.
	WeekStartsOn time.Weekday

	// FiscalYearStartMonth specifies which month the fiscal year starts.
	// Default is January (1), meaning fiscal year equals calendar year.
	// For example, if FiscalYearStartMonth is April (4), then:
	// - Fiscal Q1 runs Apr-Jun
	// - Fiscal Q2 runs Jul-Sep
	// - Fiscal Q3 runs Oct-Dec
	// - Fiscal Q4 runs Jan-Mar
	FiscalYearStartMonth time.Month
}

// NewDefaultDateResolver creates a new DefaultDateResolver with default settings.
func NewDefaultDateResolver() *DefaultDateResolver {
	return &DefaultDateResolver{
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.January,
	}
}

// NewDateResolverWithLocation creates a DefaultDateResolver with a specific timezone.
func NewDateResolverWithLocation(loc *time.Location) *DefaultDateResolver {
	return &DefaultDateResolver{
		Location:             loc,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: time.January,
	}
}

// NewDateResolverWithFiscalYear creates a DefaultDateResolver with a custom fiscal year start.
func NewDateResolverWithFiscalYear(loc *time.Location, fiscalYearStartMonth time.Month) *DefaultDateResolver {
	if fiscalYearStartMonth < time.January || fiscalYearStartMonth > time.December {
		fiscalYearStartMonth = time.January
	}
	return &DefaultDateResolver{
		Location:             loc,
		WeekStartsOn:         time.Monday,
		FiscalYearStartMonth: fiscalYearStartMonth,
	}
}

// now returns the current time in the resolver's timezone.
func (r *DefaultDateResolver) now() time.Time {
	var t time.Time
	if r.Now != nil {
		t = r.Now()
	} else {
		t = time.Now()
	}

	if r.Location != nil {
		t = t.In(r.Location)
	} else {
		t = t.UTC()
	}

	return t
}

// startOfDay returns the start of the day (00:00:00) for the given time.
func (r *DefaultDateResolver) startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// endOfDay returns the end of the day (23:59:59.999999999) for the given time.
func (r *DefaultDateResolver) endOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// startOfWeek returns the start of the week for the given time.
func (r *DefaultDateResolver) startOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	diff := int(weekday - r.WeekStartsOn)
	if diff < 0 {
		diff += 7
	}
	return r.startOfDay(t.AddDate(0, 0, -diff))
}

// startOfMonth returns the start of the month for the given time.
func (r *DefaultDateResolver) startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// startOfQuarter returns the start of the quarter for the given time.
func (r *DefaultDateResolver) startOfQuarter(t time.Time) time.Time {
	month := t.Month()
	quarterMonth := time.Month(((int(month)-1)/3)*3 + 1)
	return time.Date(t.Year(), quarterMonth, 1, 0, 0, 0, 0, t.Location())
}

// startOfYear returns the start of the year for the given time.
func (r *DefaultDateResolver) startOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
}

// startOfFiscalYear returns the start of the fiscal year for the given time.
// If FiscalYearStartMonth is January, it behaves like startOfYear.
// Otherwise, the fiscal year starts in FiscalYearStartMonth.
// For example, if fiscal year starts in April and current date is Feb 2024,
// the fiscal year started in April 2023.
func (r *DefaultDateResolver) startOfFiscalYear(t time.Time) time.Time {
	fyStartMonth := r.FiscalYearStartMonth
	if fyStartMonth == 0 {
		fyStartMonth = time.January
	}

	year := t.Year()
	// If current month is before fiscal year start, fiscal year started last calendar year
	if t.Month() < fyStartMonth {
		year--
	}

	return time.Date(year, fyStartMonth, 1, 0, 0, 0, 0, t.Location())
}

// startOfFiscalQuarter returns the start of the fiscal quarter for the given time.
// Fiscal quarters are calculated based on FiscalYearStartMonth.
func (r *DefaultDateResolver) startOfFiscalQuarter(t time.Time) time.Time {
	fyStart := r.startOfFiscalYear(t)
	fyStartMonth := int(r.FiscalYearStartMonth)
	if fyStartMonth == 0 {
		fyStartMonth = 1
	}

	// Calculate which fiscal quarter we're in
	// Months since fiscal year start (0-11)
	currentMonth := int(t.Month())
	monthsSinceFYStart := currentMonth - fyStartMonth
	if monthsSinceFYStart < 0 {
		monthsSinceFYStart += 12
	}

	// Quarter number (0-3)
	quarterNum := monthsSinceFYStart / 3

	// Start month of current fiscal quarter
	return fyStart.AddDate(0, quarterNum*3, 0)
}

// fiscalQuarterNumber returns the fiscal quarter number (1-4) for the given time.
func (r *DefaultDateResolver) fiscalQuarterNumber(t time.Time) int {
	fyStartMonth := int(r.FiscalYearStartMonth)
	if fyStartMonth == 0 {
		fyStartMonth = 1
	}

	currentMonth := int(t.Month())
	monthsSinceFYStart := currentMonth - fyStartMonth
	if monthsSinceFYStart < 0 {
		monthsSinceFYStart += 12
	}

	return (monthsSinceFYStart / 3) + 1
}

// ResolveStatic implements DateResolver.
func (r *DefaultDateResolver) ResolveStatic(ctx context.Context, literal StaticDateLiteral) (time.Time, error) {
	start, _, err := r.ResolveStaticRange(ctx, literal)
	return start, err
}

// ResolveDynamic implements DateResolver.
func (r *DefaultDateResolver) ResolveDynamic(ctx context.Context, literal *DynamicDateLiteral) (time.Time, error) {
	start, _, err := r.ResolveDynamicRange(ctx, literal)
	return start, err
}

// ResolveStaticRange implements DateResolver.
func (r *DefaultDateResolver) ResolveStaticRange(ctx context.Context, literal StaticDateLiteral) (start, end time.Time, err error) {
	now := r.now()
	today := r.startOfDay(now)

	switch literal {
	case DateToday:
		return today, r.endOfDay(today), nil

	case DateYesterday:
		yesterday := today.AddDate(0, 0, -1)
		return yesterday, r.endOfDay(yesterday), nil

	case DateTomorrow:
		tomorrow := today.AddDate(0, 0, 1)
		return tomorrow, r.endOfDay(tomorrow), nil

	case DateThisWeek:
		weekStart := r.startOfWeek(now)
		weekEnd := weekStart.AddDate(0, 0, 6)
		return weekStart, r.endOfDay(weekEnd), nil

	case DateLastWeek:
		lastWeekStart := r.startOfWeek(now).AddDate(0, 0, -7)
		lastWeekEnd := lastWeekStart.AddDate(0, 0, 6)
		return lastWeekStart, r.endOfDay(lastWeekEnd), nil

	case DateNextWeek:
		nextWeekStart := r.startOfWeek(now).AddDate(0, 0, 7)
		nextWeekEnd := nextWeekStart.AddDate(0, 0, 6)
		return nextWeekStart, r.endOfDay(nextWeekEnd), nil

	case DateThisMonth:
		monthStart := r.startOfMonth(now)
		monthEnd := monthStart.AddDate(0, 1, -1)
		return monthStart, r.endOfDay(monthEnd), nil

	case DateLastMonth:
		lastMonthStart := r.startOfMonth(now).AddDate(0, -1, 0)
		lastMonthEnd := r.startOfMonth(now).AddDate(0, 0, -1)
		return lastMonthStart, r.endOfDay(lastMonthEnd), nil

	case DateNextMonth:
		nextMonthStart := r.startOfMonth(now).AddDate(0, 1, 0)
		nextMonthEnd := r.startOfMonth(now).AddDate(0, 2, -1)
		return nextMonthStart, r.endOfDay(nextMonthEnd), nil

	case DateThisQuarter:
		quarterStart := r.startOfQuarter(now)
		quarterEnd := quarterStart.AddDate(0, 3, -1)
		return quarterStart, r.endOfDay(quarterEnd), nil

	case DateLastQuarter:
		lastQuarterStart := r.startOfQuarter(now).AddDate(0, -3, 0)
		lastQuarterEnd := r.startOfQuarter(now).AddDate(0, 0, -1)
		return lastQuarterStart, r.endOfDay(lastQuarterEnd), nil

	case DateNextQuarter:
		nextQuarterStart := r.startOfQuarter(now).AddDate(0, 3, 0)
		nextQuarterEnd := r.startOfQuarter(now).AddDate(0, 6, -1)
		return nextQuarterStart, r.endOfDay(nextQuarterEnd), nil

	case DateThisYear:
		yearStart := r.startOfYear(now)
		yearEnd := yearStart.AddDate(1, 0, -1)
		return yearStart, r.endOfDay(yearEnd), nil

	case DateLastYear:
		lastYearStart := r.startOfYear(now).AddDate(-1, 0, 0)
		lastYearEnd := r.startOfYear(now).AddDate(0, 0, -1)
		return lastYearStart, r.endOfDay(lastYearEnd), nil

	case DateNextYear:
		nextYearStart := r.startOfYear(now).AddDate(1, 0, 0)
		nextYearEnd := r.startOfYear(now).AddDate(2, 0, -1)
		return nextYearStart, r.endOfDay(nextYearEnd), nil

	case DateLast90Days:
		start := today.AddDate(0, 0, -90)
		return start, r.endOfDay(today), nil

	case DateNext90Days:
		end := today.AddDate(0, 0, 90)
		return today, r.endOfDay(end), nil

	case DateThisFiscalQuarter:
		fqStart := r.startOfFiscalQuarter(now)
		fqEnd := fqStart.AddDate(0, 3, -1)
		return fqStart, r.endOfDay(fqEnd), nil

	case DateLastFiscalQuarter:
		thisFQStart := r.startOfFiscalQuarter(now)
		lastFQStart := thisFQStart.AddDate(0, -3, 0)
		lastFQEnd := thisFQStart.AddDate(0, 0, -1)
		return lastFQStart, r.endOfDay(lastFQEnd), nil

	case DateNextFiscalQuarter:
		thisFQStart := r.startOfFiscalQuarter(now)
		nextFQStart := thisFQStart.AddDate(0, 3, 0)
		nextFQEnd := nextFQStart.AddDate(0, 3, -1)
		return nextFQStart, r.endOfDay(nextFQEnd), nil

	case DateThisFiscalYear:
		fyStart := r.startOfFiscalYear(now)
		fyEnd := fyStart.AddDate(1, 0, -1)
		return fyStart, r.endOfDay(fyEnd), nil

	case DateLastFiscalYear:
		thisFYStart := r.startOfFiscalYear(now)
		lastFYStart := thisFYStart.AddDate(-1, 0, 0)
		lastFYEnd := thisFYStart.AddDate(0, 0, -1)
		return lastFYStart, r.endOfDay(lastFYEnd), nil

	case DateNextFiscalYear:
		thisFYStart := r.startOfFiscalYear(now)
		nextFYStart := thisFYStart.AddDate(1, 0, 0)
		nextFYEnd := nextFYStart.AddDate(1, 0, -1)
		return nextFYStart, r.endOfDay(nextFYEnd), nil

	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unknown static date literal: %v", literal)
	}
}

// ResolveDynamicRange implements DateResolver.
func (r *DefaultDateResolver) ResolveDynamicRange(ctx context.Context, literal *DynamicDateLiteral) (start, end time.Time, err error) {
	if literal == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("nil dynamic date literal")
	}

	now := r.now()
	today := r.startOfDay(now)
	n := literal.N

	switch literal.Type {
	case DynamicLastNDays:
		start := today.AddDate(0, 0, -n)
		return start, r.endOfDay(today), nil

	case DynamicNextNDays:
		end := today.AddDate(0, 0, n)
		return today, r.endOfDay(end), nil

	case DynamicLastNWeeks:
		start := r.startOfWeek(now).AddDate(0, 0, -7*n)
		end := r.startOfWeek(now).AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNWeeks:
		weekEnd := r.startOfWeek(now).AddDate(0, 0, 6)
		start := weekEnd.AddDate(0, 0, 1)
		end := start.AddDate(0, 0, 7*n-1)
		return r.startOfDay(start), r.endOfDay(end), nil

	case DynamicLastNMonths:
		start := r.startOfMonth(now).AddDate(0, -n, 0)
		end := r.startOfMonth(now).AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNMonths:
		start := r.startOfMonth(now).AddDate(0, 1, 0)
		end := r.startOfMonth(now).AddDate(0, n+1, -1)
		return start, r.endOfDay(end), nil

	case DynamicLastNQuarters:
		start := r.startOfQuarter(now).AddDate(0, -3*n, 0)
		end := r.startOfQuarter(now).AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNQuarters:
		start := r.startOfQuarter(now).AddDate(0, 3, 0)
		end := r.startOfQuarter(now).AddDate(0, 3+3*n, -1)
		return start, r.endOfDay(end), nil

	case DynamicLastNYears:
		start := r.startOfYear(now).AddDate(-n, 0, 0)
		end := r.startOfYear(now).AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNYears:
		start := r.startOfYear(now).AddDate(1, 0, 0)
		end := r.startOfYear(now).AddDate(n+1, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicLastNFiscalQuarters:
		thisFQStart := r.startOfFiscalQuarter(now)
		start := thisFQStart.AddDate(0, -3*n, 0)
		end := thisFQStart.AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNFiscalQuarters:
		thisFQStart := r.startOfFiscalQuarter(now)
		start := thisFQStart.AddDate(0, 3, 0)
		end := thisFQStart.AddDate(0, 3+3*n, -1)
		return start, r.endOfDay(end), nil

	case DynamicLastNFiscalYears:
		thisFYStart := r.startOfFiscalYear(now)
		start := thisFYStart.AddDate(-n, 0, 0)
		end := thisFYStart.AddDate(0, 0, -1)
		return start, r.endOfDay(end), nil

	case DynamicNextNFiscalYears:
		thisFYStart := r.startOfFiscalYear(now)
		start := thisFYStart.AddDate(1, 0, 0)
		end := thisFYStart.AddDate(n+1, 0, -1)
		return start, r.endOfDay(end), nil

	default:
		return time.Time{}, time.Time{}, fmt.Errorf("unknown dynamic date literal type: %v", literal.Type)
	}
}

// ResolveDateParams resolves date parameters in a compiled query.
// It modifies the Params slice in place, replacing nil placeholders with resolved dates.
func ResolveDateParams(ctx context.Context, query *CompiledQuery, resolver DateResolver) error {
	if resolver == nil {
		resolver = NewDefaultDateResolver()
	}

	for _, dp := range query.DateParams {
		var resolvedValue time.Time
		var err error

		if dp.Static != nil {
			resolvedValue, err = resolver.ResolveStatic(ctx, *dp.Static)
		} else if dp.Dynamic != nil {
			resolvedValue, err = resolver.ResolveDynamic(ctx, dp.Dynamic)
		} else {
			return fmt.Errorf("invalid date parameter: neither static nor dynamic")
		}

		if err != nil {
			return fmt.Errorf("failed to resolve date parameter: %w", err)
		}

		// Update the parameter in place (0-indexed, but ParamIndex is 1-indexed)
		if dp.ParamIndex > 0 && dp.ParamIndex <= len(query.Params) {
			query.Params[dp.ParamIndex-1] = resolvedValue
		}
	}

	return nil
}

// ResolveDateParamsForRange resolves date parameters when using range comparisons.
// For literals that represent ranges (THIS_WEEK, LAST_N_DAYS), this returns
// both the start and end times for use in BETWEEN clauses.
func ResolveDateParamsForRange(ctx context.Context, query *CompiledQuery, resolver DateResolver) error {
	if resolver == nil {
		resolver = NewDefaultDateResolver()
	}

	for _, dp := range query.DateParams {
		var start, end time.Time
		var err error

		if dp.Static != nil {
			start, end, err = resolver.ResolveStaticRange(ctx, *dp.Static)
		} else if dp.Dynamic != nil {
			start, end, err = resolver.ResolveDynamicRange(ctx, dp.Dynamic)
		} else {
			return fmt.Errorf("invalid date parameter: neither static nor dynamic")
		}

		if err != nil {
			return fmt.Errorf("failed to resolve date parameter: %w", err)
		}

		// Update the start parameter
		if dp.ParamIndex > 0 && dp.ParamIndex <= len(query.Params) {
			query.Params[dp.ParamIndex-1] = start
		}

		// Update the end parameter if this is a range
		if dp.IsRange && dp.EndIndex > 0 && dp.EndIndex <= len(query.Params) {
			query.Params[dp.EndIndex-1] = end
		}
	}

	return nil
}
