package cel

import (
	"time"

	"github.com/google/uuid"
)

// RecordMap is a type alias for record data used in CEL expressions.
type RecordMap = map[string]any

// UserVars builds a CEL variable map from user context fields.
func UserVars(userID, profileID, roleID uuid.UUID) map[string]any {
	return map[string]any{
		"id":         userID.String(),
		"profile_id": profileID.String(),
		"role_id":    roleID.String(),
	}
}

// DefaultVars builds CEL variables for default expressions.
func DefaultVars(record RecordMap, user map[string]any) map[string]any {
	return map[string]any{
		"record": record,
		"user":   user,
		"now":    time.Now().UTC(),
	}
}

// ValidationVars builds CEL variables for validation rule expressions.
func ValidationVars(record, old RecordMap, user map[string]any) map[string]any {
	vars := map[string]any{
		"record": record,
		"user":   user,
		"now":    time.Now().UTC(),
	}
	if old != nil {
		vars["old"] = old
	}
	return vars
}
