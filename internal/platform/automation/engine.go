package automation

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
)

// AutomationEvent carries data from a DML operation to the automation engine.
type AutomationEvent struct {
	ObjectAPIName string
	ObjectID      uuid.UUID
	Operation     engine.Operation
	RecordIDs     []string
	Depth         int
}

// Engine evaluates and executes automation rules in response to DML events (ADR-0031).
type Engine struct {
	cache           metadata.MetadataReader
	procedureCaller ProcedureCaller
	limits          Limits
}

// ProcedureCaller invokes a named procedure with input parameters.
type ProcedureCaller interface {
	CallProcedure(ctx context.Context, code string, input map[string]any) error
}

// NewEngine creates a new automation Engine.
func NewEngine(
	cache metadata.MetadataReader,
	procedureCaller ProcedureCaller,
	limits Limits,
) *Engine {
	return &Engine{
		cache:           cache,
		procedureCaller: procedureCaller,
		limits:          limits,
	}
}

// AfterExecute is called by the DML service after a successful execute.
// It evaluates and fires matching automation rules for "after_*" events.
func (e *Engine) AfterExecute(ctx context.Context, event AutomationEvent) error {
	if event.Depth >= e.limits.MaxDepth {
		slog.Warn("automation depth limit reached",
			"object", event.ObjectAPIName,
			"depth", event.Depth,
			"max", e.limits.MaxDepth)
		return nil
	}

	eventType := mapOperationToAfterEvent(event.Operation)
	if eventType == "" {
		return nil
	}

	rules := e.getMatchingRules(event.ObjectID, eventType)
	if len(rules) == 0 {
		return nil
	}

	for _, rule := range rules {
		if err := e.executeRule(ctx, rule, event); err != nil {
			return fmt.Errorf("automationEngine.AfterExecute: rule %q: %w", rule.Name, err)
		}
	}

	return nil
}

func (e *Engine) getMatchingRules(objectID uuid.UUID, eventType string) []metadata.AutomationRule {
	allRules := e.cache.GetAutomationRules(objectID)
	var matching []metadata.AutomationRule
	for _, r := range allRules {
		if r.IsActive && r.EventType == eventType {
			matching = append(matching, r)
		}
	}
	return matching
}

func (e *Engine) executeRule(ctx context.Context, rule metadata.AutomationRule, event AutomationEvent) error {
	if e.procedureCaller == nil {
		slog.Warn("procedure caller not configured, skipping automation rule",
			"rule", rule.Name, "procedure", rule.ProcedureCode)
		return nil
	}

	input := map[string]any{
		"_recordIds":  event.RecordIDs,
		"_objectName": event.ObjectAPIName,
		"_operation":  event.Operation.String(),
	}

	return e.procedureCaller.CallProcedure(ctx, rule.ProcedureCode, input)
}

func mapOperationToAfterEvent(op engine.Operation) string {
	switch op {
	case engine.OperationInsert:
		return "after_insert"
	case engine.OperationUpdate:
		return "after_update"
	case engine.OperationDelete:
		return "after_delete"
	case engine.OperationUpsert:
		return "after_insert" // upsert fires after_insert
	default:
		return ""
	}
}
