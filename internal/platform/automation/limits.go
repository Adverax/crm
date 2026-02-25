package automation

// Limits defines platform limits for automation rule execution (ADR-0031).
type Limits struct {
	// MaxDepth is the maximum recursion depth (DML→automation→DML→...).
	MaxDepth int

	// MaxRulesPerEvent is the maximum number of active rules per object per event.
	MaxRulesPerEvent int
}

// DefaultLimits provides sensible defaults for automation limits.
var DefaultLimits = Limits{
	MaxDepth:         3,
	MaxRulesPerEvent: 20,
}
