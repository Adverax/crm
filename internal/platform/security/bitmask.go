package security

// OLS bitmask constants (CRUD).
const (
	OLSRead   = 1
	OLSCreate = 2
	OLSUpdate = 4
	OLSDelete = 8
	OLSAll    = OLSRead | OLSCreate | OLSUpdate | OLSDelete
)

// FLS bitmask constants (Read/Write).
const (
	FLSRead  = 1
	FLSWrite = 2
	FLSAll   = FLSRead | FLSWrite
)

// HasOLS checks if the given permissions include a specific OLS flag.
func HasOLS(permissions, flag int) bool {
	return permissions&flag == flag
}

// HasFLS checks if the given permissions include a specific FLS flag.
func HasFLS(permissions, flag int) bool {
	return permissions&flag == flag
}

// ComputeEffective computes the effective permissions from grant and deny sets.
// effective = OR(all grants) AND NOT OR(all denies)
func ComputeEffective(grants []int, denies []int) int {
	var grantMask int
	for _, g := range grants {
		grantMask |= g
	}

	var denyMask int
	for _, d := range denies {
		denyMask |= d
	}

	return grantMask & ^denyMask
}
