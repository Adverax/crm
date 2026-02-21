package identity

import (
	"context"

	"github.com/google/uuid"
)

// UserContext holds the security identity of the current request user.
// This type lives in a shared kernel so that packages like SOQL, DML, and
// middleware can reference it without importing the security package (ADR-0030).
type UserContext struct {
	UserID    uuid.UUID
	ProfileID uuid.UUID
	RoleID    *uuid.UUID
}

type contextKey struct{}

// ContextWithUser stores UserContext in a standard context.Context.
func ContextWithUser(ctx context.Context, uc UserContext) context.Context {
	return context.WithValue(ctx, contextKey{}, uc)
}

// UserFromContext retrieves UserContext from a standard context.Context.
// Returns the zero value and false if no UserContext is present.
func UserFromContext(ctx context.Context) (UserContext, bool) {
	uc, ok := ctx.Value(contextKey{}).(UserContext)
	return uc, ok
}
