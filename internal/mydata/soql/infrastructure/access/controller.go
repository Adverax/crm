// Package access provides access control integration for SOQL.
package access

import (
	"context"
	"errors"

	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
)

// ErrMissingDependency is returned when a required dependency is nil.
var ErrMissingDependency = errors.New("missing required dependency")

// SOQLAccessController implements engine.AccessController.
// Currently a placeholder - OLS/FLS are checked elsewhere.
type SOQLAccessController struct{}

// NewSOQLAccessController creates a new SOQLAccessController.
func NewSOQLAccessController() *SOQLAccessController {
	return &SOQLAccessController{}
}

// CanAccessObject checks if the current user can read the given object (OLS).
// Currently returns nil - OLS should be checked at a different layer.
func (c *SOQLAccessController) CanAccessObject(ctx context.Context, object string) error {
	// TODO: OLS check should be done by a service that has access to user context
	return nil
}

// CanAccessField checks if the current user can read the given field (FLS).
// Currently returns nil - FLS should be checked at a different layer.
func (c *SOQLAccessController) CanAccessField(ctx context.Context, object, field string) error {
	// TODO: FLS check should be done by a service that has access to user context
	return nil
}

// Ensure SOQLAccessController implements engine.AccessController.
// Note: RLS is enforced at the PostgreSQL level via set_config('app.user_id', ...).
var _ engine.AccessController = (*SOQLAccessController)(nil)
