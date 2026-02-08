package metadata

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/pkg/apperror"
)

const maxCompositionDepth = 2

// CompositionChecker validates composition constraints.
type CompositionChecker struct {
	// fieldsByObjectID maps object ID to its reference fields
	fieldsByObjectID map[uuid.UUID][]FieldDefinition
}

// NewCompositionChecker creates a new CompositionChecker from a list of all field definitions.
func NewCompositionChecker(fields []FieldDefinition) *CompositionChecker {
	byObj := make(map[uuid.UUID][]FieldDefinition)
	for _, f := range fields {
		if f.FieldType == FieldTypeReference {
			byObj[f.ObjectID] = append(byObj[f.ObjectID], f)
		}
	}
	return &CompositionChecker{fieldsByObjectID: byObj}
}

// ValidateNewComposition checks that adding a composition field does not violate constraints.
func (cc *CompositionChecker) ValidateNewComposition(field FieldDefinition, allObjects map[uuid.UUID]ObjectDefinition) error {
	if field.FieldType != FieldTypeReference || field.FieldSubtype == nil || *field.FieldSubtype != SubtypeComposition {
		return nil
	}

	if field.ReferencedObjectID == nil {
		return apperror.Validation("composition field must have referenced_object_id")
	}

	// Self-reference check for composition
	if *field.ReferencedObjectID == field.ObjectID {
		return apperror.Validation("composition self-reference is not allowed")
	}

	// Check composition depth: the chain from the root to the deepest leaf must be <= maxCompositionDepth.
	// The new field creates a link: field.ReferencedObjectID (parent) -> field.ObjectID (child).
	// We need to check:
	// 1. How deep is the parent in the existing chain above it (upward depth)
	// 2. How deep is the child's existing chain below it (downward depth)
	// Total depth (upward + 1 + downward) must not exceed maxCompositionDepth

	upward := cc.compositionDepthUp(*field.ReferencedObjectID, make(map[uuid.UUID]bool))
	downward := cc.compositionDepthDown(field.ObjectID, make(map[uuid.UUID]bool))

	if upward+1+downward > maxCompositionDepth {
		return apperror.Validation(fmt.Sprintf(
			"composition chain depth would exceed maximum of %d (up=%d, down=%d)",
			maxCompositionDepth, upward, downward,
		))
	}

	// Check for cycles
	if cc.wouldCreateCycle(field.ObjectID, *field.ReferencedObjectID) {
		return apperror.Validation("composition would create a cycle")
	}

	return nil
}

// compositionDepthUp counts how many composition parents are above the given object.
func (cc *CompositionChecker) compositionDepthUp(objectID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[objectID] {
		return 0
	}
	visited[objectID] = true

	maxUp := 0
	for _, f := range cc.fieldsByObjectID[objectID] {
		if f.FieldSubtype != nil && *f.FieldSubtype == SubtypeComposition && f.ReferencedObjectID != nil {
			depth := cc.compositionDepthUp(*f.ReferencedObjectID, visited)
			if depth+1 > maxUp {
				maxUp = depth + 1
			}
		}
	}
	return maxUp
}

// compositionDepthDown counts how many composition children are below the given object.
func (cc *CompositionChecker) compositionDepthDown(objectID uuid.UUID, visited map[uuid.UUID]bool) int {
	if visited[objectID] {
		return 0
	}
	visited[objectID] = true

	maxDown := 0
	for objID, fields := range cc.fieldsByObjectID {
		for _, f := range fields {
			if f.FieldSubtype != nil && *f.FieldSubtype == SubtypeComposition &&
				f.ReferencedObjectID != nil && *f.ReferencedObjectID == objectID && objID != objectID {
				depth := cc.compositionDepthDown(objID, visited)
				if depth+1 > maxDown {
					maxDown = depth + 1
				}
			}
		}
	}
	return maxDown
}

// wouldCreateCycle checks if adding a composition from childObjectID -> parentObjectID
// would create a cycle.
func (cc *CompositionChecker) wouldCreateCycle(childObjectID, parentObjectID uuid.UUID) bool {
	visited := make(map[uuid.UUID]bool)
	return cc.canReach(childObjectID, parentObjectID, visited)
}

// canReach checks if target is reachable from current via composition relationships
// by traversing downward (children of current).
func (cc *CompositionChecker) canReach(target, current uuid.UUID, visited map[uuid.UUID]bool) bool {
	if current == target {
		return true
	}
	if visited[current] {
		return false
	}
	visited[current] = true

	for objID, fields := range cc.fieldsByObjectID {
		for _, f := range fields {
			if f.FieldSubtype != nil && *f.FieldSubtype == SubtypeComposition &&
				f.ReferencedObjectID != nil && *f.ReferencedObjectID == current {
				if cc.canReach(target, objID, visited) {
					return true
				}
			}
		}
	}
	return false
}
