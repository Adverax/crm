// Package engine provides SOQL query parsing, validation, and compilation.
//
// This file defines abstractions for external dependencies to isolate the SOQL
// package from specific implementations, enabling easier testing and portability.
package engine

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// =============================================================================
// Access Control Abstraction
// =============================================================================

// AccessController checks access permissions for SOQL queries.
// Implementations can enforce Object-Level Security (OLS) and
// Field-Level Security (FLS).
// Note: Row-Level Security (RLS) is enforced at the PostgreSQL level.
type AccessController interface {
	// CanAccessObject checks if the current user can access the given object.
	// Returns nil if access is allowed, or an AccessError if denied.
	CanAccessObject(ctx context.Context, object string) error

	// CanAccessField checks if the current user can access the given field.
	// Returns nil if access is allowed, or an AccessError if denied.
	CanAccessField(ctx context.Context, object, field string) error
}

// NoopAccessController is an AccessController that allows all access.
// Useful for testing or when access control is handled elsewhere.
type NoopAccessController struct{}

// CanAccessObject implements AccessController.
func (n *NoopAccessController) CanAccessObject(ctx context.Context, object string) error {
	return nil
}

// CanAccessField implements AccessController.
func (n *NoopAccessController) CanAccessField(ctx context.Context, object, field string) error {
	return nil
}

// DenyAllAccessController is an AccessController that denies all access.
// Useful for testing error handling.
type DenyAllAccessController struct{}

// CanAccessObject implements AccessController.
func (d *DenyAllAccessController) CanAccessObject(ctx context.Context, object string) error {
	return NewAccessError(object)
}

// CanAccessField implements AccessController.
func (d *DenyAllAccessController) CanAccessField(ctx context.Context, object, field string) error {
	return NewFieldAccessError(object, field)
}

// ObjectAccessController provides object-level access control only.
// Field access is always allowed, no row filtering.
type ObjectAccessController struct {
	// AllowedObjects is the set of objects the user can access.
	// If nil, all objects are allowed.
	AllowedObjects map[string]bool
}

// CanAccessObject implements AccessController.
func (o *ObjectAccessController) CanAccessObject(ctx context.Context, object string) error {
	if o.AllowedObjects == nil {
		return nil
	}
	if !o.AllowedObjects[object] {
		return NewAccessError(object)
	}
	return nil
}

// CanAccessField implements AccessController.
func (o *ObjectAccessController) CanAccessField(ctx context.Context, object, field string) error {
	return nil
}

// FieldAccessController provides object and field level access control.
type FieldAccessController struct {
	// AllowedObjects is the set of objects the user can access.
	AllowedObjects map[string]bool

	// AllowedFields maps object names to sets of allowed field names.
	// If an object is not in this map, all fields are allowed.
	AllowedFields map[string]map[string]bool
}

// CanAccessObject implements AccessController.
func (f *FieldAccessController) CanAccessObject(ctx context.Context, object string) error {
	if f.AllowedObjects == nil {
		return nil
	}
	if !f.AllowedObjects[object] {
		return NewAccessError(object)
	}
	return nil
}

// CanAccessField implements AccessController.
func (f *FieldAccessController) CanAccessField(ctx context.Context, object, field string) error {
	if f.AllowedFields == nil {
		return nil
	}
	fields, ok := f.AllowedFields[object]
	if !ok {
		return nil // No field restrictions for this object
	}
	if !fields[field] {
		return NewFieldAccessError(object, field)
	}
	return nil
}

// FuncAccessController wraps functions as an AccessController.
type FuncAccessController struct {
	ObjectFunc func(ctx context.Context, object string) error
	FieldFunc  func(ctx context.Context, object, field string) error
}

// CanAccessObject implements AccessController.
func (f *FuncAccessController) CanAccessObject(ctx context.Context, object string) error {
	if f.ObjectFunc == nil {
		return nil
	}
	return f.ObjectFunc(ctx, object)
}

// CanAccessField implements AccessController.
func (f *FuncAccessController) CanAccessField(ctx context.Context, object, field string) error {
	if f.FieldFunc == nil {
		return nil
	}
	return f.FieldFunc(ctx, object, field)
}

// =============================================================================
// Metadata Abstraction
// =============================================================================

// MetadataProvider provides metadata about SOQL objects.
// Implementations are responsible for mapping SOQL object/field names
// to underlying database tables/columns.
type MetadataProvider interface {
	// GetObject returns metadata for a SOQL object by its name.
	// Returns nil and no error if object doesn't exist.
	GetObject(ctx context.Context, name string) (*ObjectMeta, error)

	// ListObjects returns a list of all available SOQL object names.
	ListObjects(ctx context.Context) ([]string, error)
}

// ObjectMeta describes a SOQL object (entity).
type ObjectMeta struct {
	// Name is the SOQL object name (e.g., "Account", "Contact").
	Name string

	// SchemeName is the PostgreSQL schema name (e.g., "public", "sales").
	SchemeName string

	// TableName is the underlying SQL table name (e.g., "accounts", "contacts").
	TableName string

	// Fields maps SOQL field names to their metadata.
	Fields map[string]*FieldMeta

	// Lookups maps SOQL relationship names to Child-to-Parent lookups.
	// Example: Contact has lookup "Account" that points to Account object.
	Lookups map[string]*LookupMeta

	// Relationships maps SOQL relationship names to Parent-to-Child relationships.
	// Example: Account has relationship "Contacts" that returns child Contact records.
	Relationships map[string]*RelationshipMeta
}

// QualifiedTableName returns the fully qualified and properly quoted table name.
// Uses pgx.Identifier to safely quote identifiers for PostgreSQL.
// If SchemeName is empty, returns just the quoted TableName.
func (o *ObjectMeta) QualifiedTableName() string {
	if o == nil {
		return ""
	}
	if o.SchemeName == "" {
		return pgx.Identifier{o.TableName}.Sanitize()
	}
	return pgx.Identifier{o.SchemeName, o.TableName}.Sanitize()
}

// GetField returns field metadata by SOQL field name.
// Returns nil if field doesn't exist.
func (o *ObjectMeta) GetField(name string) *FieldMeta {
	if o == nil || o.Fields == nil {
		return nil
	}
	return o.Fields[name]
}

// GetFieldByColumn returns field metadata by SQL column name.
// Returns nil if field doesn't exist.
func (o *ObjectMeta) GetFieldByColumn(column string) *FieldMeta {
	if o == nil || o.Fields == nil {
		return nil
	}
	for _, f := range o.Fields {
		if f.Column == column {
			return f
		}
	}
	return nil
}

// GetLookup returns lookup metadata by SOQL relationship name.
// Returns nil if lookup doesn't exist.
func (o *ObjectMeta) GetLookup(name string) *LookupMeta {
	if o == nil || o.Lookups == nil {
		return nil
	}
	return o.Lookups[name]
}

// GetRelationship returns relationship metadata by SOQL relationship name.
// Returns nil if relationship doesn't exist.
func (o *ObjectMeta) GetRelationship(name string) *RelationshipMeta {
	if o == nil || o.Relationships == nil {
		return nil
	}
	return o.Relationships[name]
}

// FieldMeta describes a field within a SOQL object.
type FieldMeta struct {
	// Name is the SOQL field name (e.g., "FirstName", "Email").
	Name string

	// Column is the underlying SQL column name (e.g., "first_name", "email").
	Column string

	// Type is the field's data type.
	Type FieldType

	// Nullable indicates whether the field can contain NULL values.
	Nullable bool

	// Filterable indicates whether the field can be used in WHERE clauses.
	Filterable bool

	// Sortable indicates whether the field can be used in ORDER BY clauses.
	Sortable bool

	// Groupable indicates whether the field can be used in GROUP BY clauses.
	Groupable bool

	// Aggregatable indicates whether the field can be used with aggregate functions.
	Aggregatable bool
}

// LookupMeta describes a Child-to-Parent relationship (lookup).
// Example: Contact.Account - Contact has a lookup field pointing to Account.
type LookupMeta struct {
	// Name is the SOQL relationship name used in queries (e.g., "Account").
	Name string

	// Field is the FK column in the current object's table (e.g., "account_id").
	Field string

	// TargetObject is the SOQL name of the parent object (e.g., "Account").
	TargetObject string

	// TargetField is the PK column in the parent object's table (e.g., "id").
	TargetField string
}

// RelationshipMeta describes a Parent-to-Child relationship.
// Example: Account.Contacts - Account has many Contacts.
type RelationshipMeta struct {
	// Name is the SOQL relationship name used in subqueries (e.g., "Contacts").
	Name string

	// ChildObject is the SOQL name of the child object (e.g., "Contact").
	ChildObject string

	// ChildField is the FK column in the child object's table (e.g., "account_id").
	ChildField string

	// ParentField is the PK column in the parent object's table (e.g., "id").
	ParentField string
}

// StaticMetadataProvider is a simple MetadataProvider backed by a static map.
// Useful for testing or when metadata is known at compile time.
type StaticMetadataProvider struct {
	objects map[string]*ObjectMeta
}

// NewStaticMetadataProvider creates a new StaticMetadataProvider with the given objects.
func NewStaticMetadataProvider(objects map[string]*ObjectMeta) *StaticMetadataProvider {
	return &StaticMetadataProvider{
		objects: objects,
	}
}

// GetObject implements MetadataProvider.
func (p *StaticMetadataProvider) GetObject(ctx context.Context, name string) (*ObjectMeta, error) {
	if p.objects == nil {
		return nil, nil
	}
	return p.objects[name], nil
}

// ListObjects implements MetadataProvider.
func (p *StaticMetadataProvider) ListObjects(ctx context.Context) ([]string, error) {
	if p.objects == nil {
		return nil, nil
	}
	names := make([]string, 0, len(p.objects))
	for name := range p.objects {
		names = append(names, name)
	}
	return names, nil
}

// ObjectMetaBuilder provides a fluent API for building ObjectMeta.
type ObjectMetaBuilder struct {
	meta *ObjectMeta
}

// NewObjectMeta starts building a new ObjectMeta.
func NewObjectMeta(name, schemeName, tableName string) *ObjectMetaBuilder {
	return &ObjectMetaBuilder{
		meta: &ObjectMeta{
			Name:          name,
			SchemeName:    schemeName,
			TableName:     tableName,
			Fields:        make(map[string]*FieldMeta),
			Lookups:       make(map[string]*LookupMeta),
			Relationships: make(map[string]*RelationshipMeta),
		},
	}
}

// Field adds a field to the object.
func (b *ObjectMetaBuilder) Field(name, column string, typ FieldType) *ObjectMetaBuilder {
	b.meta.Fields[name] = &FieldMeta{
		Name:         name,
		Column:       column,
		Type:         typ,
		Nullable:     true,
		Filterable:   true,
		Sortable:     true,
		Groupable:    true,
		Aggregatable: typ == FieldTypeInteger || typ == FieldTypeFloat,
	}
	return b
}

// FieldFull adds a field with full configuration.
func (b *ObjectMetaBuilder) FieldFull(field *FieldMeta) *ObjectMetaBuilder {
	b.meta.Fields[field.Name] = field
	return b
}

// Lookup adds a Child-to-Parent lookup relationship.
func (b *ObjectMetaBuilder) Lookup(name, field, targetObject, targetField string) *ObjectMetaBuilder {
	b.meta.Lookups[name] = &LookupMeta{
		Name:         name,
		Field:        field,
		TargetObject: targetObject,
		TargetField:  targetField,
	}
	return b
}

// Relationship adds a Parent-to-Child relationship.
func (b *ObjectMetaBuilder) Relationship(name, childObject, childField, parentField string) *ObjectMetaBuilder {
	b.meta.Relationships[name] = &RelationshipMeta{
		Name:        name,
		ChildObject: childObject,
		ChildField:  childField,
		ParentField: parentField,
	}
	return b
}

// Build returns the constructed ObjectMeta.
func (b *ObjectMetaBuilder) Build() *ObjectMeta {
	return b.meta
}

// =============================================================================
// Cache Abstraction
// =============================================================================

// Cache defines a generic cache interface with TTL support.
// Replaces direct dependency on proxima.crm.kernel/cache.
type Cache[K comparable, V any] interface {
	// Get retrieves a value from cache. Returns ok=false if key not found or expired.
	Get(ctx context.Context, key K) (value V, ok bool)

	// Set stores a value in cache.
	Set(ctx context.Context, key K, value V) error

	// Delete removes a specific key from cache.
	Delete(ctx context.Context, key K) error

	// Clear removes all entries from cache.
	Clear(ctx context.Context) error

	// Remove deletes entries matching the predicate.
	Remove(predicate func(V) bool) error

	// GetStats returns cache usage statistics.
	GetStats() *CacheStats
}

// CacheStats contains cache usage statistics.
type CacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
	Size   int64 `json:"size"`
}

// CacheWithGC extends Cache with garbage collection support.
type CacheWithGC[K comparable, V any] interface {
	Cache[K, V]

	// StartGarbageCollector starts background cleanup of expired entries.
	StartGarbageCollector(ctx context.Context, interval time.Duration)
}

// QueryCache is an alias for the compiled query cache type.
// Used in Dependencies to inject a pre-configured cache instance.
type QueryCache = Cache[string, *CompiledQuery]

// =============================================================================
// Pagination / Cursor Abstraction
// =============================================================================

// SortDirection defines sort order direction.
type SortDirection string

const (
	// SortAsc represents ascending sort order.
	SortAsc SortDirection = "asc"
	// SortDesc represents descending sort order.
	SortDesc SortDirection = "desc"
)

// SortKey represents a single sort column with direction.
// JSON tags are shortened for compact cursor payloads.
type SortKey struct {
	Field string        `json:"f"`
	Dir   SortDirection `json:"d"`
}

// SortKeys is a slice of SortKey with helper methods.
type SortKeys []SortKey

// Validate checks all sort keys have valid directions.
func (sk SortKeys) Validate() error {
	if len(sk) == 0 {
		return ErrEmptySortKeys
	}
	for _, key := range sk {
		if key.Dir != SortAsc && key.Dir != SortDesc {
			return ErrInvalidSortDirection
		}
	}
	return nil
}

// Fields extracts field names from sort keys.
func (sk SortKeys) Fields() []string {
	fields := make([]string, len(sk))
	for i, key := range sk {
		fields[i] = key.Field
	}
	return fields
}

// Equals compares two SortKeys slices for equality.
func (sk SortKeys) Equals(other SortKeys) bool {
	if len(sk) != len(other) {
		return false
	}
	for i, key := range sk {
		if key.Field != other[i].Field {
			return false
		}
		if key.Dir != other[i].Dir {
			return false
		}
	}
	return true
}

// Clone creates a deep copy of the sort keys.
func (sk SortKeys) Clone() SortKeys {
	if sk == nil {
		return nil
	}
	result := make(SortKeys, len(sk))
	copy(result, sk)
	return result
}

// SecretProvider provides secret key for HMAC signing of cursors.
type SecretProvider interface {
	// Secret returns the HMAC secret key.
	Secret() []byte
}

// StaticSecret is a simple SecretProvider implementation with a fixed secret.
type StaticSecret []byte

// Secret returns the static secret bytes.
func (s StaticSecret) Secret() []byte {
	return s
}

// NewStaticSecret creates a SecretProvider from a string.
func NewStaticSecret(secret string) SecretProvider {
	return StaticSecret([]byte(secret))
}

// CursorPayload contains the internal cursor structure for pagination.
type CursorPayload struct {
	Version int                    `json:"v"`
	OrderBy SortKeys               `json:"ob"`
	LastRow map[string]interface{} `json:"k"`
	FID     string                 `json:"fid"`
}

// CursorManager handles cursor encoding, decoding, and validation.
type CursorManager interface {
	// Encode creates a signed cursor string from payload.
	// Format: base64url(payload).base64url(signature)
	Encode(payload *CursorPayload) (string, error)

	// Decode parses and verifies a cursor string.
	// Returns nil, nil for empty cursor string.
	Decode(cursor string) (*CursorPayload, error)

	// ValidateContext verifies cursor matches current request context.
	ValidateContext(payload *CursorPayload, fid string, orderBy SortKeys) error

	// Next generates the next cursor string based on the last row.
	Next(lastRow map[string]interface{}, namespace string, userID int64,
		filter map[string]interface{}, orderBy SortKeys) (string, error)
}

// FIDBuilder builds Filter ID for cursor context validation.
type FIDBuilder interface {
	// BuildFID creates a unique identifier for the query context.
	// Used to detect cursor reuse across different queries.
	BuildFID(namespace string, userID int64, filter map[string]interface{}) string
}

// =============================================================================
// Pagination Errors
// =============================================================================

// PaginationError represents a pagination-related error.
type PaginationError struct {
	Code    string
	Message string
}

func (e *PaginationError) Error() string {
	return e.Message
}

var (
	// ErrEmptySortKeys indicates sort keys slice is empty.
	ErrEmptySortKeys = &PaginationError{Code: "EMPTY_SORT_KEYS", Message: "sort keys cannot be empty"}

	// ErrInvalidSortDirection indicates an invalid sort direction.
	ErrInvalidSortDirection = &PaginationError{Code: "INVALID_SORT_DIR", Message: "sort direction must be 'asc' or 'desc'"}

	// ErrInvalidCursor indicates cursor format is invalid.
	ErrInvalidCursor = &PaginationError{Code: "INVALID_CURSOR", Message: "invalid cursor format"}

	// ErrCursorTampered indicates cursor signature verification failed.
	ErrCursorTampered = &PaginationError{Code: "CURSOR_TAMPERED", Message: "cursor signature verification failed"}

	// ErrCursorMismatch indicates cursor doesn't match current query context.
	ErrCursorMismatch = &PaginationError{Code: "CURSOR_MISMATCH", Message: "cursor does not match current query context"}

	// ErrInvalidCursorVersion indicates unsupported cursor version.
	ErrInvalidCursorVersion = &PaginationError{Code: "INVALID_CURSOR_VER", Message: "unsupported cursor version"}
)

// =============================================================================
// Dependency Container
// =============================================================================

// Dependencies aggregates all external dependencies for the SOQL engine.
// This enables dependency injection and simplifies testing.
type Dependencies struct {
	// MetadataProvider provides object and field metadata.
	// Required for query validation.
	MetadataProvider MetadataProvider

	// AccessController checks object and field access permissions.
	// If nil, all access is allowed (NoopAccessController is used).
	AccessController AccessController

	// QueryCache is a pre-configured cache for compiled queries (optional).
	// If nil, query caching is disabled.
	QueryCache QueryCache

	// CursorManagerFactory creates CursorManager instances.
	// If nil, cursor-based pagination is not available.
	CursorManagerFactory func(secret SecretProvider, tieBreaker string) CursorManager

	// FIDBuilder for generating query context identifiers.
	// If nil, uses default implementation.
	FIDBuilder FIDBuilder
}

// DefaultFIDBuilder is a simple FID builder using JSON serialization.
type DefaultFIDBuilder struct{}

// BuildFID implements FIDBuilder using a simple hash of parameters.
func (b *DefaultFIDBuilder) BuildFID(namespace string, userID int64, filter map[string]interface{}) string {
	// Simple implementation - can be overridden with more sophisticated hashing
	if filter == nil {
		return namespace
	}
	return namespace // Simplified - real impl should hash filter
}
