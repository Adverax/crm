package security

import (
	"context"

	"github.com/adverax/crm/internal/pkg/identity"
)

// ContextWithUser stores UserContext in a standard context.Context.
// Delegates to the identity shared kernel (ADR-0030).
var ContextWithUser = identity.ContextWithUser

// UserFromContext retrieves UserContext from a standard context.Context.
// Delegates to the identity shared kernel (ADR-0030).
var UserFromContext = identity.UserFromContext

// Ensure compatibility: these vars have the same signatures as before.
var (
	_ func(context.Context, UserContext) context.Context = ContextWithUser
	_ func(context.Context) (UserContext, bool)          = UserFromContext
)
