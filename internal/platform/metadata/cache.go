package metadata

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// CacheLoader loads all metadata from the database for cache population.
type CacheLoader interface {
	LoadAllObjects(ctx context.Context) ([]ObjectDefinition, error)
	LoadAllFields(ctx context.Context) ([]FieldDefinition, error)
	LoadRelationships(ctx context.Context) ([]RelationshipInfo, error)
	LoadAllValidationRules(ctx context.Context) ([]ValidationRule, error)
	LoadAllFunctions(ctx context.Context) ([]Function, error)
	LoadAllObjectViews(ctx context.Context) ([]ObjectView, error)
	LoadAllProcedures(ctx context.Context) ([]Procedure, error)
	RefreshMaterializedView(ctx context.Context) error
}

// MetadataReader provides read-only access to cached metadata (ADR-0030).
// All consumers should depend on this interface, not *MetadataCache.
type MetadataReader interface {
	GetObjectByID(id uuid.UUID) (ObjectDefinition, bool)
	GetObjectByAPIName(apiName string) (ObjectDefinition, bool)
	GetFieldByID(id uuid.UUID) (FieldDefinition, bool)
	GetFieldsByObjectID(objectID uuid.UUID) []FieldDefinition
	GetForwardRelationships(childObjectID uuid.UUID) []RelationshipInfo
	GetReverseRelationships(parentObjectID uuid.UUID) []RelationshipInfo
	ListObjectAPINames() []string
	GetValidationRules(objectID uuid.UUID) []ValidationRule
	GetFunctions() []Function
	GetFunctionByName(name string) (Function, bool)
	GetObjectViews(objectID uuid.UUID) []ObjectView
	GetProcedureByCode(code string) (Procedure, bool)
	GetProcedures() []Procedure
}

// MetadataCache is an in-memory cache of metadata backed by a PostgreSQL materialized view
// for the relationship registry (ADR-0006).
type MetadataCache struct {
	mu sync.RWMutex

	objectsByID      map[uuid.UUID]ObjectDefinition
	objectsByAPIName map[string]ObjectDefinition
	fieldsByID       map[uuid.UUID]FieldDefinition
	fieldsByObjectID map[uuid.UUID][]FieldDefinition

	// Relationship registry (from MV)
	forwardRels map[uuid.UUID][]RelationshipInfo // child_object_id → []rel
	reverseRels map[uuid.UUID][]RelationshipInfo // parent_object_id → []rel

	// Validation rules
	validationRulesByObjectID map[uuid.UUID][]ValidationRule

	// Custom functions
	functionsByName map[string]Function

	// Object views (ADR-0022)
	objectViewsByObjectID map[uuid.UUID][]ObjectView

	// Procedures (ADR-0024)
	proceduresByCode map[string]Procedure

	loader CacheLoader
	loaded bool
}

// NewMetadataCache creates a new MetadataCache.
func NewMetadataCache(loader CacheLoader) *MetadataCache {
	return &MetadataCache{
		objectsByID:               make(map[uuid.UUID]ObjectDefinition),
		objectsByAPIName:          make(map[string]ObjectDefinition),
		fieldsByID:                make(map[uuid.UUID]FieldDefinition),
		fieldsByObjectID:          make(map[uuid.UUID][]FieldDefinition),
		forwardRels:               make(map[uuid.UUID][]RelationshipInfo),
		reverseRels:               make(map[uuid.UUID][]RelationshipInfo),
		validationRulesByObjectID: make(map[uuid.UUID][]ValidationRule),
		functionsByName:           make(map[string]Function),
		objectViewsByObjectID:     make(map[uuid.UUID][]ObjectView),
		proceduresByCode:          make(map[string]Procedure),
		loader:                    loader,
	}
}

// Load performs a full cache rebuild from the database.
func (c *MetadataCache) Load(ctx context.Context) error {
	objects, err := c.loader.LoadAllObjects(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: objects: %w", err)
	}

	fields, err := c.loader.LoadAllFields(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: fields: %w", err)
	}

	rels, err := c.loader.LoadRelationships(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: relationships: %w", err)
	}

	rules, err := c.loader.LoadAllValidationRules(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: validation rules: %w", err)
	}

	functions, err := c.loader.LoadAllFunctions(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: functions: %w", err)
	}

	objectViews, err := c.loader.LoadAllObjectViews(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: object views: %w", err)
	}

	procedures, err := c.loader.LoadAllProcedures(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.Load: procedures: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.objectsByID = make(map[uuid.UUID]ObjectDefinition, len(objects))
	c.objectsByAPIName = make(map[string]ObjectDefinition, len(objects))
	for _, obj := range objects {
		c.objectsByID[obj.ID] = obj
		c.objectsByAPIName[obj.APIName] = obj
	}

	c.fieldsByID = make(map[uuid.UUID]FieldDefinition, len(fields))
	c.fieldsByObjectID = make(map[uuid.UUID][]FieldDefinition)
	for _, f := range fields {
		c.fieldsByID[f.ID] = f
		c.fieldsByObjectID[f.ObjectID] = append(c.fieldsByObjectID[f.ObjectID], f)
	}

	c.forwardRels = make(map[uuid.UUID][]RelationshipInfo)
	c.reverseRels = make(map[uuid.UUID][]RelationshipInfo)
	for _, rel := range rels {
		c.forwardRels[rel.ChildObjectID] = append(c.forwardRels[rel.ChildObjectID], rel)
		if rel.ParentObjectID != uuid.Nil {
			c.reverseRels[rel.ParentObjectID] = append(c.reverseRels[rel.ParentObjectID], rel)
		}
	}

	c.validationRulesByObjectID = make(map[uuid.UUID][]ValidationRule)
	for _, rule := range rules {
		c.validationRulesByObjectID[rule.ObjectID] = append(c.validationRulesByObjectID[rule.ObjectID], rule)
	}

	c.functionsByName = make(map[string]Function, len(functions))
	for _, fn := range functions {
		c.functionsByName[fn.Name] = fn
	}

	c.objectViewsByObjectID = make(map[uuid.UUID][]ObjectView)
	for _, ov := range objectViews {
		c.objectViewsByObjectID[ov.ObjectID] = append(c.objectViewsByObjectID[ov.ObjectID], ov)
	}

	c.proceduresByCode = make(map[string]Procedure, len(procedures))
	for _, p := range procedures {
		c.proceduresByCode[p.Code] = p
	}

	c.loaded = true
	return nil
}

// Invalidate refreshes the materialized view and rebuilds the in-memory cache.
func (c *MetadataCache) Invalidate(ctx context.Context) error {
	if err := c.loader.RefreshMaterializedView(ctx); err != nil {
		return fmt.Errorf("metadataCache.Invalidate: refresh MV: %w", err)
	}
	return c.Load(ctx)
}

// GetObjectByID returns an object definition by ID from cache.
func (c *MetadataCache) GetObjectByID(id uuid.UUID) (ObjectDefinition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	obj, ok := c.objectsByID[id]
	return obj, ok
}

// GetObjectByAPIName returns an object definition by API name from cache.
func (c *MetadataCache) GetObjectByAPIName(apiName string) (ObjectDefinition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	obj, ok := c.objectsByAPIName[apiName]
	return obj, ok
}

// GetFieldByID returns a field definition by ID from cache.
func (c *MetadataCache) GetFieldByID(id uuid.UUID) (FieldDefinition, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	f, ok := c.fieldsByID[id]
	return f, ok
}

// GetFieldsByObjectID returns all field definitions for an object from cache.
func (c *MetadataCache) GetFieldsByObjectID(objectID uuid.UUID) []FieldDefinition {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.fieldsByObjectID[objectID]
}

// GetForwardRelationships returns relationships where the given object is the child (has FK).
func (c *MetadataCache) GetForwardRelationships(childObjectID uuid.UUID) []RelationshipInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.forwardRels[childObjectID]
}

// GetReverseRelationships returns relationships where the given object is the parent (referenced).
func (c *MetadataCache) GetReverseRelationships(parentObjectID uuid.UUID) []RelationshipInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.reverseRels[parentObjectID]
}

// ListObjectAPINames returns all object API names in the cache.
func (c *MetadataCache) ListObjectAPINames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	names := make([]string, 0, len(c.objectsByAPIName))
	for name := range c.objectsByAPIName {
		names = append(names, name)
	}
	return names
}

// GetValidationRules returns all validation rules for an object.
func (c *MetadataCache) GetValidationRules(objectID uuid.UUID) []ValidationRule {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.validationRulesByObjectID[objectID]
}

// LoadValidationRules reloads only validation rules into the cache.
func (c *MetadataCache) LoadValidationRules(ctx context.Context) error {
	rules, err := c.loader.LoadAllValidationRules(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.LoadValidationRules: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.validationRulesByObjectID = make(map[uuid.UUID][]ValidationRule)
	for _, rule := range rules {
		c.validationRulesByObjectID[rule.ObjectID] = append(c.validationRulesByObjectID[rule.ObjectID], rule)
	}
	return nil
}

// GetFunctions returns all cached custom functions.
func (c *MetadataCache) GetFunctions() []Function {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Function, 0, len(c.functionsByName))
	for _, fn := range c.functionsByName {
		result = append(result, fn)
	}
	return result
}

// GetFunctionByName returns a custom function by name from cache.
func (c *MetadataCache) GetFunctionByName(name string) (Function, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	fn, ok := c.functionsByName[name]
	return fn, ok
}

// LoadFunctions reloads only custom functions into the cache.
func (c *MetadataCache) LoadFunctions(ctx context.Context) error {
	functions, err := c.loader.LoadAllFunctions(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.LoadFunctions: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.functionsByName = make(map[string]Function, len(functions))
	for _, fn := range functions {
		c.functionsByName[fn.Name] = fn
	}
	return nil
}

// GetObjectViews returns all object views for an object from cache.
func (c *MetadataCache) GetObjectViews(objectID uuid.UUID) []ObjectView {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.objectViewsByObjectID[objectID]
}

// LoadObjectViews reloads only object views into the cache.
func (c *MetadataCache) LoadObjectViews(ctx context.Context) error {
	views, err := c.loader.LoadAllObjectViews(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.LoadObjectViews: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.objectViewsByObjectID = make(map[uuid.UUID][]ObjectView)
	for _, ov := range views {
		c.objectViewsByObjectID[ov.ObjectID] = append(c.objectViewsByObjectID[ov.ObjectID], ov)
	}
	return nil
}

// GetProcedureByCode returns a procedure by code from cache.
func (c *MetadataCache) GetProcedureByCode(code string) (Procedure, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.proceduresByCode[code]
	return p, ok
}

// GetProcedures returns all cached procedures.
func (c *MetadataCache) GetProcedures() []Procedure {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Procedure, 0, len(c.proceduresByCode))
	for _, p := range c.proceduresByCode {
		result = append(result, p)
	}
	return result
}

// LoadProcedures reloads only procedures into the cache.
func (c *MetadataCache) LoadProcedures(ctx context.Context) error {
	procedures, err := c.loader.LoadAllProcedures(ctx)
	if err != nil {
		return fmt.Errorf("metadataCache.LoadProcedures: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.proceduresByCode = make(map[string]Procedure, len(procedures))
	for _, p := range procedures {
		c.proceduresByCode[p.Code] = p
	}
	return nil
}

// IsLoaded returns whether the cache has been loaded.
func (c *MetadataCache) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}
