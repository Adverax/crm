package engine

import (
	"context"
	"fmt"
)

// ValidatedDML represents a validated DML statement with resolved metadata.
type ValidatedDML struct {
	AST       *DMLStatement
	Object    *ObjectMeta
	Operation Operation

	// For INSERT/UPSERT
	Fields       []*FieldMeta // Resolved field metadata in order
	RowCount     int          // Number of rows in VALUES
	ValuesPerRow int          // Number of values per row

	// For UPDATE
	Assignments []*ValidatedAssignment // Resolved assignments

	// For UPSERT
	ExternalIdField *FieldMeta // The external ID field for conflict resolution

	// For UPDATE/DELETE
	HasWhere bool // Whether statement has WHERE clause
}

// ValidatedAssignment represents a validated field assignment.
type ValidatedAssignment struct {
	Field *FieldMeta
	Value *Expr
}

// Validator validates DML statements against metadata and access rules.
type Validator struct {
	metadata MetadataProvider
	access   WriteAccessController
	limits   *Limits
}

// NewValidator creates a new Validator.
func NewValidator(metadata MetadataProvider, access WriteAccessController, limits *Limits) *Validator {
	if limits == nil {
		limits = &DefaultLimits
	}
	if access == nil {
		access = &NoopWriteAccessController{}
	}
	return &Validator{
		metadata: metadata,
		access:   access,
		limits:   limits,
	}
}

// Validate validates a parsed DML statement.
func (v *Validator) Validate(ctx context.Context, ast *DMLStatement) (*ValidatedDML, error) {
	switch {
	case ast.Insert != nil:
		return v.validateInsert(ctx, ast, ast.Insert)
	case ast.Update != nil:
		return v.validateUpdate(ctx, ast, ast.Update)
	case ast.Delete != nil:
		return v.validateDelete(ctx, ast, ast.Delete)
	case ast.Upsert != nil:
		return v.validateUpsert(ctx, ast, ast.Upsert)
	default:
		return nil, NewValidationError(ErrCodeInvalidExpression, "empty DML statement")
	}
}

// validateInsert validates an INSERT statement.
func (v *Validator) validateInsert(ctx context.Context, ast *DMLStatement, ins *InsertStatement) (*ValidatedDML, error) {
	// Get object metadata
	obj, err := v.metadata.GetObject(ctx, ins.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	if obj == nil {
		return nil, UnknownObjectError(ins.Object)
	}

	// Check object write access
	if err := v.access.CanWriteObject(ctx, ins.Object, OperationInsert); err != nil {
		return nil, err
	}

	// Validate fields exist and are not duplicates
	fields, err := v.resolveFields(obj, ins.Fields)
	if err != nil {
		return nil, err
	}

	// Check field write access
	if err := v.access.CheckWritableFields(ctx, ins.Object, ins.Fields); err != nil {
		return nil, err
	}

	// Check fields are not read-only
	for _, f := range fields {
		if f.ReadOnly {
			return nil, ReadOnlyFieldError(ins.Object, f.Name)
		}
		if f.Calculated {
			return nil, ReadOnlyFieldError(ins.Object, f.Name)
		}
	}

	// Validate VALUES
	if len(ins.Values) == 0 {
		return nil, NewValidationError(ErrCodeInvalidValue, "INSERT requires at least one VALUES row")
	}

	// Check batch size limit
	if err := v.limits.CheckBatchSize(len(ins.Values)); err != nil {
		return nil, err
	}

	// Check fields per row limit
	if err := v.limits.CheckFieldsPerRow(len(ins.Fields)); err != nil {
		return nil, err
	}

	// Validate each row has correct number of values
	for i, row := range ins.Values {
		if len(row.Values) != len(ins.Fields) {
			return nil, NewValidationError(ErrCodeInvalidValue,
				fmt.Sprintf("row %d has %d values, expected %d", i+1, len(row.Values), len(ins.Fields)))
		}

		// Validate value types
		for j, val := range row.Values {
			if err := v.validateExpr(ctx, obj, fields[j], val); err != nil {
				return nil, err
			}
		}
	}

	// Check required fields are provided
	requiredFields := obj.GetRequiredFields()
	fieldSet := make(map[string]bool)
	for _, f := range fields {
		fieldSet[f.Name] = true
	}
	for _, req := range requiredFields {
		if !fieldSet[req.Name] {
			return nil, MissingRequiredFieldError(ins.Object, req.Name)
		}
	}

	return &ValidatedDML{
		AST:          ast,
		Object:       obj,
		Operation:    OperationInsert,
		Fields:       fields,
		RowCount:     len(ins.Values),
		ValuesPerRow: len(ins.Fields),
	}, nil
}

// validateUpdate validates an UPDATE statement.
func (v *Validator) validateUpdate(ctx context.Context, ast *DMLStatement, upd *UpdateStatement) (*ValidatedDML, error) {
	// Get object metadata
	obj, err := v.metadata.GetObject(ctx, upd.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	if obj == nil {
		return nil, UnknownObjectError(upd.Object)
	}

	// Check object write access
	if err := v.access.CanWriteObject(ctx, upd.Object, OperationUpdate); err != nil {
		return nil, err
	}

	// Check WHERE requirement
	hasWhere := upd.Where != nil
	if v.limits.RequireWhereOnUpdate && !hasWhere {
		return nil, NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("UPDATE on %s requires a WHERE clause", upd.Object))
	}

	// Validate assignments
	if len(upd.Assignments) == 0 {
		return nil, NewValidationError(ErrCodeInvalidExpression, "UPDATE requires at least one assignment")
	}

	// Check for duplicate fields in SET
	seenFields := make(map[string]bool)
	var fieldNames []string
	for _, assign := range upd.Assignments {
		if seenFields[assign.Field] {
			return nil, DuplicateFieldError(upd.Object, assign.Field)
		}
		seenFields[assign.Field] = true
		fieldNames = append(fieldNames, assign.Field)
	}

	// Check field write access
	if err := v.access.CheckWritableFields(ctx, upd.Object, fieldNames); err != nil {
		return nil, err
	}

	// Validate each assignment
	assignments := make([]*ValidatedAssignment, len(upd.Assignments))
	for i, assign := range upd.Assignments {
		field := obj.GetField(assign.Field)
		if field == nil {
			return nil, UnknownFieldError(upd.Object, assign.Field)
		}
		if field.ReadOnly {
			return nil, ReadOnlyFieldError(upd.Object, assign.Field)
		}
		if field.Calculated {
			return nil, ReadOnlyFieldError(upd.Object, assign.Field)
		}

		// Validate value expression
		if err := v.validateExpr(ctx, obj, field, assign.Value); err != nil {
			return nil, err
		}

		assignments[i] = &ValidatedAssignment{
			Field: field,
			Value: assign.Value,
		}
	}

	// Validate WHERE clause if present
	if upd.Where != nil {
		if err := v.validateExpression(ctx, obj, upd.Where); err != nil {
			return nil, fmt.Errorf("invalid WHERE clause: %w", err)
		}
	}

	return &ValidatedDML{
		AST:         ast,
		Object:      obj,
		Operation:   OperationUpdate,
		Assignments: assignments,
		HasWhere:    hasWhere,
	}, nil
}

// validateDelete validates a DELETE statement.
func (v *Validator) validateDelete(ctx context.Context, ast *DMLStatement, del *DeleteStatement) (*ValidatedDML, error) {
	// Get object metadata
	obj, err := v.metadata.GetObject(ctx, del.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	if obj == nil {
		return nil, UnknownObjectError(del.Object)
	}

	// Check object write access
	if err := v.access.CanWriteObject(ctx, del.Object, OperationDelete); err != nil {
		return nil, err
	}

	// Check WHERE requirement (safety check)
	hasWhere := del.Where != nil
	if v.limits.RequireWhereOnDelete && !hasWhere {
		return nil, DeleteRequiresWhereError(del.Object)
	}

	// Validate WHERE clause if present
	if del.Where != nil {
		if err := v.validateExpression(ctx, obj, del.Where); err != nil {
			return nil, fmt.Errorf("invalid WHERE clause: %w", err)
		}
	}

	return &ValidatedDML{
		AST:       ast,
		Object:    obj,
		Operation: OperationDelete,
		HasWhere:  hasWhere,
	}, nil
}

// validateUpsert validates an UPSERT statement.
func (v *Validator) validateUpsert(ctx context.Context, ast *DMLStatement, ups *UpsertStatement) (*ValidatedDML, error) {
	// Get object metadata
	obj, err := v.metadata.GetObject(ctx, ups.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	if obj == nil {
		return nil, UnknownObjectError(ups.Object)
	}

	// Check object write access (UPSERT needs both INSERT and UPDATE rights)
	if err := v.access.CanWriteObject(ctx, ups.Object, OperationUpsert); err != nil {
		return nil, err
	}

	// Validate external ID field
	extField := obj.GetField(ups.ExternalIdField)
	if extField == nil {
		return nil, ExternalIdNotFoundError(ups.Object, ups.ExternalIdField)
	}
	if !extField.IsExternalId && !extField.IsUnique {
		return nil, ExternalIdNotFoundError(ups.Object, ups.ExternalIdField)
	}

	// Validate external ID field is in the field list
	extFieldInList := false
	for _, f := range ups.Fields {
		if f == ups.ExternalIdField {
			extFieldInList = true
			break
		}
	}
	if !extFieldInList {
		return nil, NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("external ID field %s must be included in field list", ups.ExternalIdField))
	}

	// Validate fields exist and are not duplicates
	fields, err := v.resolveFields(obj, ups.Fields)
	if err != nil {
		return nil, err
	}

	// Check field write access
	if err := v.access.CheckWritableFields(ctx, ups.Object, ups.Fields); err != nil {
		return nil, err
	}

	// Check fields are not read-only (except external ID can be unique but writable)
	for _, f := range fields {
		if f.ReadOnly {
			return nil, ReadOnlyFieldError(ups.Object, f.Name)
		}
		if f.Calculated {
			return nil, ReadOnlyFieldError(ups.Object, f.Name)
		}
	}

	// Validate VALUES
	if len(ups.Values) == 0 {
		return nil, NewValidationError(ErrCodeInvalidValue, "UPSERT requires at least one VALUES row")
	}

	// Check batch size limit
	if err := v.limits.CheckBatchSize(len(ups.Values)); err != nil {
		return nil, err
	}

	// Check fields per row limit
	if err := v.limits.CheckFieldsPerRow(len(ups.Fields)); err != nil {
		return nil, err
	}

	// Validate each row has correct number of values
	for i, row := range ups.Values {
		if len(row.Values) != len(ups.Fields) {
			return nil, NewValidationError(ErrCodeInvalidValue,
				fmt.Sprintf("row %d has %d values, expected %d", i+1, len(row.Values), len(ups.Fields)))
		}

		// Validate value types
		for j, val := range row.Values {
			if err := v.validateExpr(ctx, obj, fields[j], val); err != nil {
				return nil, err
			}
		}
	}

	return &ValidatedDML{
		AST:             ast,
		Object:          obj,
		Operation:       OperationUpsert,
		Fields:          fields,
		RowCount:        len(ups.Values),
		ValuesPerRow:    len(ups.Fields),
		ExternalIdField: extField,
	}, nil
}

// resolveFields resolves field names to FieldMeta and checks for duplicates.
func (v *Validator) resolveFields(obj *ObjectMeta, fieldNames []string) ([]*FieldMeta, error) {
	seen := make(map[string]bool)
	fields := make([]*FieldMeta, len(fieldNames))

	for i, name := range fieldNames {
		if seen[name] {
			return nil, DuplicateFieldError(obj.Name, name)
		}
		seen[name] = true

		field := obj.GetField(name)
		if field == nil {
			return nil, UnknownFieldError(obj.Name, name)
		}
		fields[i] = field
	}

	return fields, nil
}

// validateExpr validates an expression in VALUES or SET and returns its inferred type.
func (v *Validator) validateExpr(ctx context.Context, obj *ObjectMeta, field *FieldMeta, expr *Expr) error {
	if expr == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty expression")
	}

	switch {
	case expr.FuncCall != nil:
		return v.validateFuncCall(ctx, obj, field, expr.FuncCall)
	case expr.Const != nil:
		return v.validateConstType(obj, field, expr.Const)
	case expr.Field != nil:
		return v.validateFieldRef(ctx, obj, expr.Field)
	default:
		return NewValidationError(ErrCodeInvalidExpression, "invalid expression")
	}
}

// validateFuncCall validates a function call expression.
func (v *Validator) validateFuncCall(ctx context.Context, obj *ObjectMeta, field *FieldMeta, fc *FuncCall) error {
	// Check argument count
	minArgs := fc.Name.MinArgs()
	maxArgs := fc.Name.MaxArgs()

	if len(fc.Args) < minArgs {
		return NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("function %s requires at least %d argument(s), got %d", fc.Name, minArgs, len(fc.Args)))
	}
	if maxArgs >= 0 && len(fc.Args) > maxArgs {
		return NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("function %s accepts at most %d argument(s), got %d", fc.Name, maxArgs, len(fc.Args)))
	}

	// Validate each argument (recursively)
	argTypes := make([]FieldType, len(fc.Args))
	for i, arg := range fc.Args {
		// For function arguments, we don't enforce field type matching - just validate the expression
		if err := v.validateExprArg(ctx, obj, arg); err != nil {
			return err
		}
		argTypes[i] = arg.FieldType
	}

	// Set the result type
	fc.FieldType = fc.Name.ResultType(argTypes)

	// If we have a target field, check result type compatibility
	if field != nil && fc.FieldType != FieldTypeUnknown && fc.FieldType != FieldTypeNull {
		if !fc.FieldType.IsCompatibleWith(field.Type) {
			return TypeMismatchError(obj.Name, field.Name, field.Type, fc.FieldType)
		}
	}

	return nil
}

// validateExprArg validates an expression used as a function argument.
func (v *Validator) validateExprArg(ctx context.Context, obj *ObjectMeta, expr *Expr) error {
	if expr == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty expression")
	}

	switch {
	case expr.FuncCall != nil:
		// Recursive validation for nested function calls
		if err := v.validateFuncCall(ctx, obj, nil, expr.FuncCall); err != nil {
			return err
		}
		expr.FieldType = expr.FuncCall.FieldType
	case expr.Const != nil:
		expr.FieldType = expr.Const.GetFieldType()
	case expr.Field != nil:
		if err := v.validateFieldRef(ctx, obj, expr.Field); err != nil {
			return err
		}
		expr.FieldType = expr.Field.FieldType
	default:
		return NewValidationError(ErrCodeInvalidExpression, "invalid expression")
	}

	return nil
}

// validateFieldRef validates a field reference in an expression.
func (v *Validator) validateFieldRef(ctx context.Context, obj *ObjectMeta, field *Field) error {
	if field == nil || field.Name == "" {
		return NewValidationError(ErrCodeInvalidExpression, "empty field reference")
	}

	fieldMeta := obj.GetField(field.Name)
	if fieldMeta == nil {
		return UnknownFieldError(obj.Name, field.Name)
	}

	field.FieldType = fieldMeta.Type
	return nil
}

// validateConstType validates that a constant value is compatible with a field type.
func (v *Validator) validateConstType(obj *ObjectMeta, field *FieldMeta, val *Const) error {
	valType := val.GetFieldType()

	// NULL is always allowed for nullable fields
	if valType == FieldTypeNull {
		if !field.Nullable && field.Required {
			return NewValidationError(ErrCodeInvalidValue,
				fmt.Sprintf("NULL not allowed for required field %s.%s", obj.Name, field.Name))
		}
		return nil
	}

	// Check type compatibility
	if !valType.IsCompatibleWith(field.Type) {
		return TypeMismatchError(obj.Name, field.Name, field.Type, valType)
	}

	return nil
}

// validateExpression validates a WHERE expression.
func (v *Validator) validateExpression(ctx context.Context, obj *ObjectMeta, expr *Expression) error {
	if expr == nil || expr.Or == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty expression")
	}
	return v.validateOrExpr(ctx, obj, expr.Or)
}

func (v *Validator) validateOrExpr(ctx context.Context, obj *ObjectMeta, or *OrExpr) error {
	for _, and := range or.And {
		if err := v.validateAndExpr(ctx, obj, and); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateAndExpr(ctx context.Context, obj *ObjectMeta, and *AndExpr) error {
	for _, not := range and.Not {
		if err := v.validateNotExpr(ctx, obj, not); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateNotExpr(ctx context.Context, obj *ObjectMeta, not *NotExpr) error {
	return v.validateCompareExpr(ctx, obj, not.Compare)
}

func (v *Validator) validateCompareExpr(ctx context.Context, obj *ObjectMeta, cmp *CompareExpr) error {
	if cmp.Left == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty comparison expression")
	}

	if err := v.validateInExpr(ctx, obj, cmp.Left); err != nil {
		return err
	}

	if cmp.Right != nil {
		if err := v.validateInExpr(ctx, obj, cmp.Right); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateInExpr(ctx context.Context, obj *ObjectMeta, in *InExpr) error {
	if err := v.validateLikeExpr(ctx, obj, in.Left); err != nil {
		return err
	}

	if in.In {
		for _, val := range in.Values {
			if err := v.validateValue(ctx, obj, val); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateLikeExpr(ctx context.Context, obj *ObjectMeta, like *LikeExpr) error {
	if err := v.validateIsExpr(ctx, obj, like.Left); err != nil {
		return err
	}

	if like.Pattern != nil {
		if err := v.validateValue(ctx, obj, like.Pattern); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateIsExpr(ctx context.Context, obj *ObjectMeta, is *IsExpr) error {
	return v.validatePrimary(ctx, obj, is.Left)
}

func (v *Validator) validatePrimary(ctx context.Context, obj *ObjectMeta, primary *Primary) error {
	if primary == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty primary expression")
	}

	switch {
	case primary.Subexpression != nil:
		return v.validateExpression(ctx, obj, primary.Subexpression)
	case primary.Const != nil:
		// Constants are always valid
		primary.Const.GetFieldType()
		return nil
	case primary.Field != nil:
		return v.validateField(ctx, obj, primary.Field)
	default:
		return NewValidationError(ErrCodeInvalidExpression, "invalid primary expression")
	}
}

func (v *Validator) validateValue(ctx context.Context, obj *ObjectMeta, val *Value) error {
	if val == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty value")
	}

	if val.Const != nil {
		val.Const.GetFieldType()
		return nil
	}

	if val.Field != nil {
		return v.validateField(ctx, obj, val.Field)
	}

	return NewValidationError(ErrCodeInvalidExpression, "invalid value")
}

func (v *Validator) validateField(ctx context.Context, obj *ObjectMeta, field *Field) error {
	if field == nil || field.Name == "" {
		return NewValidationError(ErrCodeInvalidExpression, "empty field reference")
	}

	fieldMeta := obj.GetField(field.Name)
	if fieldMeta == nil {
		return UnknownFieldError(obj.Name, field.Name)
	}

	field.FieldType = fieldMeta.Type
	return nil
}
