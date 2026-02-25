package automation

import (
	"context"
	"fmt"

	"github.com/adverax/crm/internal/platform/dml/engine"
	"github.com/adverax/crm/internal/platform/metadata"
)

// DMLPostExecuteHook adapts the automation Engine to the DML PostExecuteHook interface.
type DMLPostExecuteHook struct {
	automationEngine *Engine
	cache            metadata.MetadataReader
}

// NewDMLPostExecuteHook creates a new DMLPostExecuteHook.
func NewDMLPostExecuteHook(automationEngine *Engine, cache metadata.MetadataReader) *DMLPostExecuteHook {
	return &DMLPostExecuteHook{
		automationEngine: automationEngine,
		cache:            cache,
	}
}

// AfterDMLExecute is called after a successful DML execute. It converts
// the compiled DML + result into an AutomationEvent and delegates to the engine.
func (h *DMLPostExecuteHook) AfterDMLExecute(ctx context.Context, compiled *engine.CompiledDML, result *engine.Result) error {
	if compiled == nil || result == nil {
		return nil
	}

	objectAPIName := compiled.Object
	obj, ok := h.cache.GetObjectByAPIName(objectAPIName)
	if !ok {
		return nil
	}

	var recordIDs []string
	switch compiled.Operation {
	case engine.OperationInsert, engine.OperationUpsert:
		recordIDs = result.InsertedIds
	case engine.OperationUpdate:
		recordIDs = result.UpdatedIds
	case engine.OperationDelete:
		recordIDs = result.DeletedIds
	}

	if len(recordIDs) == 0 {
		return nil
	}

	event := AutomationEvent{
		ObjectAPIName: objectAPIName,
		ObjectID:      obj.ID,
		Operation:     compiled.Operation,
		RecordIDs:     recordIDs,
		Depth:         0,
	}

	if err := h.automationEngine.AfterExecute(ctx, event); err != nil {
		return fmt.Errorf("dmlPostExecuteHook: %w", err)
	}

	return nil
}
