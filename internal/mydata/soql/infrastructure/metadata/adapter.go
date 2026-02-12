// Package metadata provides an adapter that bridges platform metadata to SOQL engine.
package metadata

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/proxima-research/proxima.crm.kernel/cache"
	"github.com/proxima-research/proxima.crm.platform/internal/data/soql/application/engine"
	metaModel "github.com/proxima-research/proxima.crm.platform/internal/metadata/domain"
)

// ObjectRepository defines the interface for accessing object metadata.
// This matches the subset of metadataService.ObjectRepository needed by SOQL.
type ObjectRepository interface {
	GetObjectDescription(ctx context.Context, objectApiName metaModel.ObjectApiName, params metaModel.ObjectDescriptionParams) (*metaModel.ObjectDescription, error)
	GetListObjects(ctx context.Context) ([]*metaModel.ObjectDefinition, error)
}

// DefaultCacheTTL is the default time-to-live for cached metadata entries.
const DefaultCacheTTL = 5 * time.Minute

// MetadataAdapter adapts platform metadata service to SOQL engine.MetadataProvider.
// It includes a thread-safe cache with lazy loading for improved performance.
type MetadataAdapter struct {
	repo ObjectRepository
	ttl  time.Duration

	// objectCache stores object metadata keyed by lowercase object name
	objectCache cache.GenericCache[string, *engine.ObjectMeta]

	// objectListCache stores the list of object names
	objectListCache cache.GenericCache[string, []string]

	// loadMu prevents concurrent loads of the same object
	loadMu sync.Mutex
}

// NewMetadataAdapter creates a new MetadataAdapter with default cache TTL.
func NewMetadataAdapter(repo ObjectRepository) *MetadataAdapter {
	return NewMetadataAdapterWithTTL(repo, DefaultCacheTTL)
}

// NewMetadataAdapterWithTTL creates a new MetadataAdapter with custom cache TTL.
func NewMetadataAdapterWithTTL(repo ObjectRepository, ttl time.Duration) *MetadataAdapter {
	return &MetadataAdapter{
		repo:            repo,
		ttl:             ttl,
		objectCache:     cache.New[string, *engine.ObjectMeta](ttl),
		objectListCache: cache.New[string, []string](ttl),
	}
}

// StartGarbageCollector starts background cleanup of expired cache entries.
func (a *MetadataAdapter) StartGarbageCollector(ctx context.Context, interval time.Duration) {
	a.objectCache.(*cache.MemoryGenericCache[string, *engine.ObjectMeta]).StartGarbageCollector(ctx, interval)
	a.objectListCache.(*cache.MemoryGenericCache[string, []string]).StartGarbageCollector(ctx, interval)
}

// GetObject implements engine.MetadataProvider with caching.
func (a *MetadataAdapter) GetObject(ctx context.Context, name string) (*engine.ObjectMeta, error) {
	// Normalize name for cache key (case-insensitive)
	cacheKey := strings.ToLower(name)

	// Try to get from cache
	if obj, found := a.objectCache.Get(ctx, cacheKey); found {
		return obj, nil
	}

	// Cache miss - load from repository with mutex to prevent duplicate loads
	return a.loadAndCache(ctx, name, cacheKey)
}

// loadAndCache loads object from repository and caches it.
func (a *MetadataAdapter) loadAndCache(ctx context.Context, name, cacheKey string) (*engine.ObjectMeta, error) {
	a.loadMu.Lock()
	defer a.loadMu.Unlock()

	// Double-check after acquiring lock
	if obj, found := a.objectCache.Get(ctx, cacheKey); found {
		return obj, nil
	}

	// Load from repository
	desc, err := a.repo.GetObjectDescription(ctx, metaModel.ObjectApiName(name), metaModel.ObjectDescriptionParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object description for %s: %w", name, err)
	}

	var obj *engine.ObjectMeta
	if desc != nil {
		obj = a.convertObjectMeta(desc)
	}

	// Cache the result (including nil for negative caching)
	_ = a.objectCache.Set(ctx, cacheKey, obj)

	return obj, nil
}

// ListObjects implements engine.MetadataProvider with caching.
func (a *MetadataAdapter) ListObjects(ctx context.Context) ([]string, error) {
	const cacheKey = "_list_"

	// Try to get from cache
	if list, found := a.objectListCache.Get(ctx, cacheKey); found {
		// Return a copy to prevent external modification
		result := make([]string, len(list))
		copy(result, list)
		return result, nil
	}

	// Cache miss - load from repository
	a.loadMu.Lock()
	defer a.loadMu.Unlock()

	// Double-check after acquiring lock
	if list, found := a.objectListCache.Get(ctx, cacheKey); found {
		result := make([]string, len(list))
		copy(result, list)
		return result, nil
	}

	objects, err := a.repo.GetListObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	names := make([]string, 0, len(objects))
	for _, obj := range objects {
		names = append(names, string(obj.ApiName))
	}

	// Cache the result
	_ = a.objectListCache.Set(ctx, cacheKey, names)

	// Return a copy
	result := make([]string, len(names))
	copy(result, names)
	return result, nil
}

// InvalidateObject removes a specific object from the cache.
func (a *MetadataAdapter) InvalidateObject(ctx context.Context, name string) {
	cacheKey := strings.ToLower(name)
	_ = a.objectCache.Delete(ctx, cacheKey)
}

// InvalidateAll clears the entire cache.
func (a *MetadataAdapter) InvalidateAll(ctx context.Context) {
	_ = a.objectCache.Clear(ctx)
	_ = a.objectListCache.Clear(ctx)
}

// CacheStats returns cache statistics.
type CacheStats struct {
	ObjectCacheStats     *cache.Stats
	ObjectListCacheStats *cache.Stats
}

// Stats returns current cache statistics.
func (a *MetadataAdapter) Stats() CacheStats {
	return CacheStats{
		ObjectCacheStats:     a.objectCache.GetStats(),
		ObjectListCacheStats: a.objectListCache.GetStats(),
	}
}

// convertObjectMeta converts metaModel.ObjectDescription to engine.ObjectMeta.
func (a *MetadataAdapter) convertObjectMeta(desc *metaModel.ObjectDescription) *engine.ObjectMeta {
	meta := &engine.ObjectMeta{
		Name:          string(desc.ApiName),
		SchemeName:    a.schemaName(desc),
		TableName:     a.baseTableName(desc),
		Fields:        make(map[string]*engine.FieldMeta),
		Lookups:       make(map[string]*engine.LookupMeta),
		Relationships: make(map[string]*engine.RelationshipMeta),
	}

	// Add system fields
	a.addSystemFields(meta)

	// Convert fields
	for _, field := range desc.Fields {
		fieldMeta := a.convertFieldMeta(field)
		meta.Fields[string(field.ApiName)] = fieldMeta

		// If field is a reference, add lookup
		if field.Type == metaModel.FieldTypeReference {
			lookup := a.createLookupFromReference(field)
			if lookup != nil {
				meta.Lookups[lookup.Name] = lookup
			}
		}
	}

	return meta
}

// schemaName returns the PostgreSQL schema name for an object.
func (a *MetadataAdapter) schemaName(desc *metaModel.ObjectDescription) string {
	// TODO: add logic for different object types if needed
	return "dataview"
}

// baseTableName returns the SQL table name (without schema) for an object.
func (a *MetadataAdapter) baseTableName(desc *metaModel.ObjectDescription) string {
	return strings.ToLower(string(desc.ApiName))
}

// addSystemFields adds standard system fields that exist on all objects.
func (a *MetadataAdapter) addSystemFields(meta *engine.ObjectMeta) {
	// todo: adjust system fields as per platform conventions

	// Id field - primary key
	meta.Fields["Id"] = &engine.FieldMeta{
		Name:         "Id",
		Column:       "record_id",
		Type:         engine.FieldTypeID,
		Nullable:     false,
		Filterable:   true,
		Sortable:     true,
		Groupable:    true,
		Aggregatable: false,
	}

	// CreatedAt
	meta.Fields["CreatedAt"] = &engine.FieldMeta{
		Name:         "CreatedAt",
		Column:       "created_at",
		Type:         engine.FieldTypeDateTime,
		Nullable:     false,
		Filterable:   true,
		Sortable:     true,
		Groupable:    false,
		Aggregatable: false,
	}

	// UpdatedAt
	meta.Fields["UpdatedAt"] = &engine.FieldMeta{
		Name:         "UpdatedAt",
		Column:       "updated_at",
		Type:         engine.FieldTypeDateTime,
		Nullable:     false,
		Filterable:   true,
		Sortable:     true,
		Groupable:    false,
		Aggregatable: false,
	}

	// CreatedBy
	meta.Fields["CreatedBy"] = &engine.FieldMeta{
		Name:         "CreatedBy",
		Column:       "created_by",
		Type:         engine.FieldTypeID,
		Nullable:     false,
		Filterable:   true,
		Sortable:     true,
		Groupable:    true,
		Aggregatable: false,
	}

	// UpdatedBy
	meta.Fields["UpdatedBy"] = &engine.FieldMeta{
		Name:         "UpdatedBy",
		Column:       "updated_by",
		Type:         engine.FieldTypeID,
		Nullable:     true,
		Filterable:   true,
		Sortable:     true,
		Groupable:    true,
		Aggregatable: false,
	}

	// OwnerId
	meta.Fields["OwnerId"] = &engine.FieldMeta{
		Name:         "OwnerId",
		Column:       "owner_id",
		Type:         engine.FieldTypeID,
		Nullable:     false,
		Filterable:   true,
		Sortable:     true,
		Groupable:    true,
		Aggregatable: false,
	}
}

// convertFieldMeta converts metaModel.FieldDefinition to engine.FieldMeta.
func (a *MetadataAdapter) convertFieldMeta(field *metaModel.FieldDefinition) *engine.FieldMeta {
	return &engine.FieldMeta{
		Name:         string(field.ApiName),
		Column:       strings.ToLower(string(field.ApiName)),
		Type:         a.mapFieldType(field.Type, field.Subtype),
		Nullable:     !field.Required,
		Filterable:   true,
		Sortable:     a.isSortable(field.Type),
		Groupable:    a.isGroupable(field.Type),
		Aggregatable: a.isAggregatable(field.Type, field.Subtype),
	}
}

// mapFieldType maps metaModel.FieldType to engine.FieldType.
func (a *MetadataAdapter) mapFieldType(fieldType metaModel.FieldType, subtype metaModel.FieldSubtype) engine.FieldType {
	switch fieldType {
	case metaModel.FieldTypeText:
		return engine.FieldTypeString
	case metaModel.FieldTypeNumber:
		if subtype == metaModel.FieldSubtypeInteger {
			return engine.FieldTypeInteger
		}
		return engine.FieldTypeFloat
	case metaModel.FieldTypeBoolean:
		return engine.FieldTypeBoolean
	case metaModel.FieldTypeDateTime:
		if subtype == metaModel.FieldSubtypeDate {
			return engine.FieldTypeDate
		}
		return engine.FieldTypeDateTime
	case metaModel.FieldTypeReference:
		return engine.FieldTypeID
	case metaModel.FieldTypePicklist:
		return engine.FieldTypeString
	case metaModel.FieldTypeFormula:
		return engine.FieldTypeString // Formula result type varies, default to string
	case metaModel.FieldTypeFile:
		return engine.FieldTypeString // File fields store URLs/paths
	default:
		return engine.FieldTypeString
	}
}

// isSortable returns whether a field type is sortable.
func (a *MetadataAdapter) isSortable(fieldType metaModel.FieldType) bool {
	switch fieldType {
	case metaModel.FieldTypeFile:
		return false
	default:
		return true
	}
}

// isGroupable returns whether a field type is groupable.
func (a *MetadataAdapter) isGroupable(fieldType metaModel.FieldType) bool {
	switch fieldType {
	case metaModel.FieldTypeFile:
		return false
	default:
		return true
	}
}

// isAggregatable returns whether a field type is aggregatable.
func (a *MetadataAdapter) isAggregatable(fieldType metaModel.FieldType, subtype metaModel.FieldSubtype) bool {
	if fieldType == metaModel.FieldTypeNumber {
		return true
	}
	return false
}

// createLookupFromReference creates a LookupMeta from a reference field.
func (a *MetadataAdapter) createLookupFromReference(field *metaModel.FieldDefinition) *engine.LookupMeta {
	// Reference fields have config with target object info
	if field.Config == nil {
		return nil
	}

	// Get target object from config
	targetObj, ok := field.Config["referenceObject"].(string)
	if !ok {
		return nil
	}

	// Lookup name is field name without "Id" suffix (e.g., "AccountId" -> "Account")
	lookupName := string(field.ApiName)
	if strings.HasSuffix(lookupName, "Id") {
		lookupName = strings.TrimSuffix(lookupName, "Id")
	}

	return &engine.LookupMeta{
		Name:         lookupName,
		Field:        strings.ToLower(string(field.ApiName)),
		TargetObject: targetObj,
		TargetField:  "record_id",
	}
}

// Ensure MetadataAdapter implements engine.MetadataProvider.
var _ engine.MetadataProvider = (*MetadataAdapter)(nil)
