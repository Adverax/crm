package metadata

import (
	"time"

	"github.com/google/uuid"
)

// ProfileNavigation represents a per-profile sidebar configuration (ADR-0032).
type ProfileNavigation struct {
	ID        uuid.UUID `json:"id"`
	ProfileID uuid.UUID `json:"profile_id"`
	Config    NavConfig `json:"config"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NavConfig holds the sidebar navigation structure stored as JSONB.
type NavConfig struct {
	Groups []NavGroup `json:"groups"`
}

// NavGroup represents a collapsible group of navigation items.
type NavGroup struct {
	Key   string    `json:"key"`
	Label string    `json:"label"`
	Icon  string    `json:"icon,omitempty"`
	Items []NavItem `json:"items"`
}

// NavItem represents a single entry in a navigation group.
// Type: "object" | "link" | "divider" | "page".
type NavItem struct {
	Type          string `json:"type"`
	ObjectAPIName string `json:"object_api_name,omitempty"`
	OVAPIName     string `json:"ov_api_name,omitempty"`
	Label         string `json:"label,omitempty"`
	URL           string `json:"url,omitempty"`
	Icon          string `json:"icon,omitempty"`
}

// CreateProfileNavigationInput is the input for creating a navigation config.
type CreateProfileNavigationInput struct {
	ProfileID uuid.UUID
	Config    NavConfig
}

// UpdateProfileNavigationInput is the input for updating a navigation config.
type UpdateProfileNavigationInput struct {
	Config NavConfig
}
