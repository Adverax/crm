package metadata

import (
	"time"

	"github.com/google/uuid"
)

// ProfileDashboard represents a per-profile dashboard configuration (ADR-0032).
type ProfileDashboard struct {
	ID        uuid.UUID       `json:"id"`
	ProfileID uuid.UUID       `json:"profile_id"`
	Config    DashboardConfig `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// DashboardConfig holds the dashboard widget layout stored as JSONB.
type DashboardConfig struct {
	Widgets []DashboardWidget `json:"widgets"`
}

// DashboardWidget represents a single widget on the dashboard.
// Type: "list" | "metric" | "link_list".
type DashboardWidget struct {
	Key           string     `json:"key"`
	Type          string     `json:"type"`
	Label         string     `json:"label"`
	Size          string     `json:"size"`
	Query         string     `json:"query,omitempty"`
	Columns       []string   `json:"columns,omitempty"`
	ObjectAPIName string     `json:"object_api_name,omitempty"`
	Format        string     `json:"format,omitempty"`
	Links         []DashLink `json:"links,omitempty"`
}

// DashLink represents a link item in a link_list widget.
type DashLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Icon  string `json:"icon,omitempty"`
}

// CreateProfileDashboardInput is the input for creating a dashboard config.
type CreateProfileDashboardInput struct {
	ProfileID uuid.UUID
	Config    DashboardConfig
}

// UpdateProfileDashboardInput is the input for updating a dashboard config.
type UpdateProfileDashboardInput struct {
	Config DashboardConfig
}
