package dml

// DescribeAdapter implements DMLTargetExtractor interface for DescribeHandler.
type DescribeAdapter struct{}

// NewDescribeAdapter creates a new DescribeAdapter.
func NewDescribeAdapter() *DescribeAdapter {
	return &DescribeAdapter{}
}

// ExtractTargets delegates to the package-level ExtractTargets function.
func (a *DescribeAdapter) ExtractTargets(statements []string) []DMLTargetInfo {
	return ExtractTargets(statements)
}
