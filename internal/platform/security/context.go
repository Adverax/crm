package security

import "context"

type contextKey struct{}

// ContextWithUser stores UserContext in a standard context.Context.
// This allows non-Gin code (SOQL/DML engines) to access user identity.
func ContextWithUser(ctx context.Context, uc UserContext) context.Context {
	return context.WithValue(ctx, contextKey{}, uc)
}

// UserFromContext retrieves UserContext from a standard context.Context.
// Returns the zero value and false if no UserContext is present.
func UserFromContext(ctx context.Context) (UserContext, bool) {
	uc, ok := ctx.Value(contextKey{}).(UserContext)
	return uc, ok
}
