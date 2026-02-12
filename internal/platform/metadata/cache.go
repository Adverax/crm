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
	RefreshMaterializedView(ctx context.Context) error
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

	loader CacheLoader
	loaded bool
}

// NewMetadataCache creates a new MetadataCache.
func NewMetadataCache(loader CacheLoader) *MetadataCache {
	return &MetadataCache{
		objectsByID:      make(map[uuid.UUID]ObjectDefinition),
		objectsByAPIName: make(map[string]ObjectDefinition),
		fieldsByID:       make(map[uuid.UUID]FieldDefinition),
		fieldsByObjectID: make(map[uuid.UUID][]FieldDefinition),
		forwardRels:      make(map[uuid.UUID][]RelationshipInfo),
		reverseRels:      make(map[uuid.UUID][]RelationshipInfo),
		loader:           loader,
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

// IsLoaded returns whether the cache has been loaded.
func (c *MetadataCache) IsLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.loaded
}
