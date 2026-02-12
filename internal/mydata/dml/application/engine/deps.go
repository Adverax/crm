// Package engine provides DML statement parsing, validation, and compilation.
//
// This file defines abstractions for external dependencies to isolate the DML
// package from specific implementations, enabling easier testing and portability.
package engine

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// =============================================================================
// Write Access Control Abstraction
// =============================================================================

// Operation represents a DML operation type.
type Operation int

const (
	OperationInsert Operation = iota
	OperationUpdate
	OperationDelete
	OperationUpsert
)

func (op Operation) String() string {
	switch op {
	case OperationInsert:
		return "INSERT"
	case OperationUpdate:
		return "UPDATE"
	case OperationDelete:
		return "DELETE"
	case OperationUpsert:
		return "UPSERT"
	default:
		return "UNKNOWN"
	}
}

// WriteAccessController checks write permissions for DML operations.
// Implementations enforce Object-Level Security (OLS) and
// Field-Level Security (FLS) for write operations.
type WriteAccessController interface {
	// CanWriteObject checks if the current user can perform the given operation on the object.
	// Returns nil if access is allowed, or an AccessError if denied.
	CanWriteObject(ctx context.Context, object string, op Operation) error

	// CheckWritableFields checks if the current user can write to the given fields.
	// Returns nil if all fields are writable, or an AccessError listing non-writable fields.
	CheckWritableFields(ctx context.Context, object string, fields []string) error
}

// NoopWriteAccessController is a WriteAccessController that allows all write access.
// Useful for testing or when access control is handled elsewhere.
type NoopWriteAccessController struct{}

// CanWriteObject implements WriteAccessController.
func (n *NoopWriteAccessController) CanWriteObject(ctx context.Context, object string, op Operation) error {
	return nil
}

// CheckWritableFields implements WriteAccessController.
func (n *NoopWriteAccessController) CheckWritableFields(ctx context.Context, object string, fields []string) error {
	return nil
}

// DenyAllWriteAccessController is a WriteAccessController that denies all access.
// Useful for testing error handling.
type DenyAllWriteAccessController struct{}

// CanWriteObject implements WriteAccessController.
func (d *DenyAllWriteAccessController) CanWriteObject(ctx context.Context, object string, op Operation) error {
	return NewWriteAccessError(object, op)
}

// CheckWritableFields implements WriteAccessController.
func (d *DenyAllWriteAccessController) CheckWritableFields(ctx context.Context, object string, fields []string) error {
	if len(fields) > 0 {
		return NewFieldWriteAccessError(object, fields[0])
	}
	return nil
}

// FuncWriteAccessController wraps functions as a WriteAccessController.
type FuncWriteAccessController struct {
	ObjectFunc func(ctx context.Context, object string, op Operation) error
	FieldFunc  func(ctx context.Context, object string, fields []string) error
}

// CanWriteObject implements WriteAccessController.
func (f *FuncWriteAccessController) CanWriteObject(ctx context.Context, object string, op Operation) error {
	if f.ObjectFunc == nil {
		return nil
	}
	return f.ObjectFunc(ctx, object, op)
}

// CheckWritableFields implements WriteAccessController.
func (f *FuncWriteAccessController) CheckWritableFields(ctx context.Context, object string, fields []string) error {
	if f.FieldFunc == nil {
		return nil
	}
	return f.FieldFunc(ctx, object, fields)
}

// =============================================================================
// Metadata Abstraction
// =============================================================================

// MetadataProvider provides metadata about DML objects.
// Implementations are responsible for mapping DML object/field names
// to underlying database tables/columns.
type MetadataProvider interface {
	// GetObject returns metadata for a DML object by its name.
	// Returns nil and no error if object doesn't exist.
	GetObject(ctx context.Context, name string) (*ObjectMeta, error)
}

// ObjectMeta describes a DML object (entity).
type ObjectMeta struct {
	// Name is the DML object name (e.g., "Account", "Contact").
	Name string

	// SchemaName is the database schema name (e.g., "public", "sales").
	// If empty, defaults to "public".
	SchemaName string

	// TableName is the underlying SQL table name without schema (e.g., "accounts", "contacts").
	TableName string

	// Fields maps DML field names to their metadata.
	Fields map[string]*FieldMeta

	// PrimaryKey is the name of the primary key field (usually "Id" or "record_id").
	PrimaryKey string
}

// Table returns the fully qualified table name (schema.table).
func (o *ObjectMeta) Table() string {
	if o == nil {
		return ""
	}
	schema := o.SchemaName
	if schema == "" {
		schema = "public"
	}
	return schema + "." + o.TableName
}

// GetField returns field metadata by DML field name.
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

// GetWritableFields returns all writable fields.
func (o *ObjectMeta) GetWritableFields() []*FieldMeta {
	if o == nil || o.Fields == nil {
		return nil
	}
	var result []*FieldMeta
	for _, f := range o.Fields {
		if !f.ReadOnly && !f.Calculated {
			result = append(result, f)
		}
	}
	return result
}

// GetRequiredFields returns all required (non-nullable) fields.
func (o *ObjectMeta) GetRequiredFields() []*FieldMeta {
	if o == nil || o.Fields == nil {
		return nil
	}
	var result []*FieldMeta
	for _, f := range o.Fields {
		if f.Required && !f.HasDefault {
			result = append(result, f)
		}
	}
	return result
}

// FieldMeta describes a field within a DML object.
type FieldMeta struct {
	// Name is the DML field name (e.g., "FirstName", "Email").
	Name string

	// Column is the underlying SQL column name (e.g., "first_name", "email").
	Column string

	// Type is the field's data type.
	Type FieldType

	// Nullable indicates whether the field can contain NULL values.
	Nullable bool

	// Required indicates whether the field must be provided on INSERT.
	// A field is required if it's not nullable and has no default value.
	Required bool

	// ReadOnly indicates whether the field cannot be modified (system fields).
	ReadOnly bool

	// Calculated indicates whether the field is computed (cannot be written directly).
	Calculated bool

	// HasDefault indicates whether the field has a database default value.
	HasDefault bool

	// IsExternalId indicates whether the field can be used as external ID for UPSERT.
	IsExternalId bool

	// IsUnique indicates whether the field has a unique constraint.
	IsUnique bool
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

// ObjectMetaBuilder provides a fluent API for building ObjectMeta.
type ObjectMetaBuilder struct {
	meta *ObjectMeta
}

// NewObjectMeta starts building a new ObjectMeta.
// tableName is the table name without schema prefix.
func NewObjectMeta(name, tableName string) *ObjectMetaBuilder {
	return &ObjectMetaBuilder{
		meta: &ObjectMeta{
			Name:       name,
			SchemaName: "public",
			TableName:  tableName,
			Fields:     make(map[string]*FieldMeta),
			PrimaryKey: "record_id",
		},
	}
}

// Schema sets the schema name (default is "public").
func (b *ObjectMetaBuilder) Schema(schema string) *ObjectMetaBuilder {
	b.meta.SchemaName = schema
	return b
}

// PrimaryKey sets the primary key field name.
func (b *ObjectMetaBuilder) PrimaryKey(pk string) *ObjectMetaBuilder {
	b.meta.PrimaryKey = pk
	return b
}

// Field adds a writable field to the object.
func (b *ObjectMetaBuilder) Field(name, column string, typ FieldType) *ObjectMetaBuilder {
	b.meta.Fields[name] = &FieldMeta{
		Name:     name,
		Column:   column,
		Type:     typ,
		Nullable: true,
	}
	return b
}

// RequiredField adds a required field to the object.
func (b *ObjectMetaBuilder) RequiredField(name, column string, typ FieldType) *ObjectMetaBuilder {
	b.meta.Fields[name] = &FieldMeta{
		Name:     name,
		Column:   column,
		Type:     typ,
		Nullable: false,
		Required: true,
	}
	return b
}

// ReadOnlyField adds a read-only field to the object.
func (b *ObjectMetaBuilder) ReadOnlyField(name, column string, typ FieldType) *ObjectMetaBuilder {
	b.meta.Fields[name] = &FieldMeta{
		Name:     name,
		Column:   column,
		Type:     typ,
		Nullable: true,
		ReadOnly: true,
	}
	return b
}

// ExternalIdField adds a field that can be used for UPSERT operations.
func (b *ObjectMetaBuilder) ExternalIdField(name, column string, typ FieldType) *ObjectMetaBuilder {
	b.meta.Fields[name] = &FieldMeta{
		Name:         name,
		Column:       column,
		Type:         typ,
		Nullable:     false,
		IsExternalId: true,
		IsUnique:     true,
	}
	return b
}

// FieldFull adds a field with full configuration.
func (b *ObjectMetaBuilder) FieldFull(field *FieldMeta) *ObjectMetaBuilder {
	b.meta.Fields[field.Name] = field
	return b
}

// Build returns the constructed ObjectMeta.
func (b *ObjectMetaBuilder) Build() *ObjectMeta {
	return b.meta
}

// =============================================================================
// Database Abstraction
// =============================================================================

// DB is the interface for database operations.
// Compatible with pgx.Conn, pgx.Pool, and pgx.Tx.
type DB interface {
	// Exec executes a SQL statement and returns the command tag.
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)

	// Query executes a query and returns rows.
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)

	// QueryRow executes a query that returns at most one row.
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// =============================================================================
// Executor Abstraction
// =============================================================================

// Executor executes compiled DML statements.
type Executor interface {
	// Execute executes a compiled DML statement and returns the result.
	Execute(ctx context.Context, compiled *CompiledDML) (*Result, error)
}

// Result represents the result of a DML operation.
type Result struct {
	// RowsAffected is the number of rows affected by the operation.
	RowsAffected int64

	// InsertedIds contains the IDs of inserted records (for INSERT/UPSERT).
	InsertedIds []string

	// UpdatedIds contains the IDs of updated records (for UPDATE/UPSERT).
	UpdatedIds []string

	// DeletedIds contains the IDs of deleted records (for DELETE).
	DeletedIds []string
}

// DefaultExecutor is a simple executor that uses a DB interface.
type DefaultExecutor struct {
	db DB
}

// NewDefaultExecutor creates a new DefaultExecutor.
func NewDefaultExecutor(db DB) *DefaultExecutor {
	return &DefaultExecutor{db: db}
}

// Execute implements Executor.
func (e *DefaultExecutor) Execute(ctx context.Context, compiled *CompiledDML) (*Result, error) {
	rows, err := e.db.Query(ctx, compiled.SQL, compiled.Params...)
	if err != nil {
		return nil, NewExecutionError("failed to execute DML", err)
	}
	defer rows.Close()

	result := &Result{}

	// Collect returned IDs
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, NewExecutionError("failed to scan result", err)
		}

		switch compiled.Operation {
		case OperationInsert:
			result.InsertedIds = append(result.InsertedIds, id)
		case OperationUpdate:
			result.UpdatedIds = append(result.UpdatedIds, id)
		case OperationDelete:
			result.DeletedIds = append(result.DeletedIds, id)
		case OperationUpsert:
			// For UPSERT, we can't easily distinguish between insert and update
			result.InsertedIds = append(result.InsertedIds, id)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, NewExecutionError("error reading results", err)
	}

	// Set RowsAffected based on collected IDs
	switch compiled.Operation {
	case OperationInsert, OperationUpsert:
		result.RowsAffected = int64(len(result.InsertedIds))
	case OperationUpdate:
		result.RowsAffected = int64(len(result.UpdatedIds))
	case OperationDelete:
		result.RowsAffected = int64(len(result.DeletedIds))
	}

	return result, nil
}

// =============================================================================
// Dependency Container
// =============================================================================

// Dependencies aggregates all external dependencies for the DML engine.
// This enables dependency injection and simplifies testing.
type Dependencies struct {
	// MetadataProvider provides object and field metadata.
	// Required for statement validation.
	MetadataProvider MetadataProvider

	// WriteAccessController checks object and field write permissions.
	// If nil, all write access is allowed (NoopWriteAccessController is used).
	WriteAccessController WriteAccessController

	// Executor executes compiled DML statements.
	// If nil, statements can only be compiled, not executed.
	Executor Executor
}
