package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2/lexer"
)

// ValidatedQuery represents a validated SOQL query with resolved metadata.
type ValidatedQuery struct {
	AST               *Grammar
	RootObject        *ObjectMeta
	ResolvedRefs      map[string]*ResolvedRef // path -> resolved reference
	Subqueries        []*ValidatedSubquery
	WhereSubqueries   []*ValidatedWhereSubquery
	TypeofExpressions []*ValidatedTypeof
	FieldCount        int
}

// ResolvedRef represents a resolved field reference.
type ResolvedRef struct {
	Path       []string    // Original path from AST
	Object     *ObjectMeta // Target object
	Field      *FieldMeta  // Target field
	Joins      []*Join     // Required joins to reach this field
	LookupPath []string    // Objects traversed via lookups
}

// Join represents a join required to resolve a lookup.
type Join struct {
	FromObject *ObjectMeta
	FromField  string
	ToObject   *ObjectMeta
	ToField    string
	Alias      string
}

// ValidatedSubquery represents a validated Parent-to-Child subquery.
type ValidatedSubquery struct {
	Relationship *RelationshipMeta
	ChildObject  *ObjectMeta
	AST          *RelationshipSubquery
	FieldCount   int
}

// ValidatedWhereSubquery represents a validated WHERE IN subquery (semi-join).
type ValidatedWhereSubquery struct {
	AST          *WhereSubquery
	Object       *ObjectMeta
	Field        *FieldMeta              // The single selected field
	ResolvedRefs map[string]*ResolvedRef // Resolved field references in the subquery
}

// ValidatedTypeof represents a validated TYPEOF expression.
type ValidatedTypeof struct {
	AST         *TypeofExpression
	Field       string                 // Polymorphic field name
	WhenClauses []*ValidatedTypeofWhen // Validated WHEN clauses
	ElseFields  []*FieldMeta           // Fields for ELSE clause
}

// ValidatedTypeofWhen represents a validated WHEN clause in TYPEOF.
type ValidatedTypeofWhen struct {
	ObjectType string       // Object type name
	Object     *ObjectMeta  // Resolved object metadata
	Fields     []*FieldMeta // Resolved fields
}

// Validator validates SOQL AST against metadata and access rules.
type Validator struct {
	metadata MetadataProvider
	access   AccessController
	limits   *Limits
}

// NewValidator creates a new Validator.
func NewValidator(metadata MetadataProvider, access AccessController, limits *Limits) *Validator {
	if limits == nil {
		limits = &DefaultLimits
	}
	if access == nil {
		access = &NoopAccessController{}
	}
	return &Validator{
		metadata: metadata,
		access:   access,
		limits:   limits,
	}
}

// validationContext holds state during validation.
type validationContext struct {
	ctx                  context.Context
	validator            *Validator
	rootObject           *ObjectMeta
	resolvedRefs         map[string]*ResolvedRef
	subqueries           []*ValidatedSubquery
	whereSubqueries      []*ValidatedWhereSubquery
	typeofExpressions    []*ValidatedTypeof
	fieldCount           int
	lookupDepth          int
	inSubquery           bool
	inWhereSubquery      bool
	parentCtx            *validationContext // For nested subqueries
	withSecurityEnforced bool               // WITH SECURITY_ENFORCED flag
}

func newValidationContext(ctx context.Context, v *Validator, root *ObjectMeta, withSecurityEnforced bool) *validationContext {
	return &validationContext{
		ctx:                  ctx,
		validator:            v,
		rootObject:           root,
		resolvedRefs:         make(map[string]*ResolvedRef),
		subqueries:           nil,
		whereSubqueries:      nil,
		fieldCount:           0,
		lookupDepth:          0,
		inSubquery:           false,
		inWhereSubquery:      false,
		parentCtx:            nil,
		withSecurityEnforced: withSecurityEnforced,
	}
}

// Validate validates a parsed SOQL query.
func (v *Validator) Validate(ctx context.Context, ast *Grammar) (*ValidatedQuery, error) {
	// Check query length
	// Note: We don't have access to original query string here,
	// this should be checked before parsing

	// Validate FROM clause - get root object
	rootObject, err := v.metadata.GetObject(ctx, ast.From)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}
	if rootObject == nil {
		return nil, UnknownObjectError(ast.From)
	}

	// Check object access
	if err := v.access.CanAccessObject(ctx, ast.From); err != nil {
		return nil, err
	}

	vctx := newValidationContext(ctx, v, rootObject, ast.WithSecurityEnforced)

	// Validate SELECT clause
	if err := v.validateSelect(vctx, ast.Select); err != nil {
		return nil, err
	}

	// Validate WHERE clause
	if ast.Where != nil {
		if err := v.validateExpression(vctx, ast.Where); err != nil {
			return nil, fmt.Errorf("invalid WHERE clause: %w", err)
		}
	}

	// Validate GROUP BY clause
	if err := v.validateGroupBy(vctx, ast.GroupBy); err != nil {
		return nil, err
	}

	// Validate HAVING clause
	if ast.Having != nil {
		if err := v.validateExpression(vctx, ast.Having); err != nil {
			return nil, fmt.Errorf("invalid HAVING clause: %w", err)
		}
	}

	// Validate ORDER BY clause
	if err := v.validateOrderBy(vctx, ast.OrderBy); err != nil {
		return nil, err
	}

	// Validate LIMIT
	if ast.Limit != nil {
		if err := v.limits.CheckRecords(*ast.Limit); err != nil {
			return nil, err
		}
	}

	// Validate OFFSET
	if ast.Offset != nil {
		if err := v.limits.CheckOffset(*ast.Offset); err != nil {
			return nil, err
		}
	}

	// Check subquery count
	if err := v.limits.CheckSubqueries(len(vctx.subqueries)); err != nil {
		return nil, err
	}

	return &ValidatedQuery{
		AST:               ast,
		RootObject:        rootObject,
		ResolvedRefs:      vctx.resolvedRefs,
		Subqueries:        vctx.subqueries,
		WhereSubqueries:   vctx.whereSubqueries,
		TypeofExpressions: vctx.typeofExpressions,
		FieldCount:        vctx.fieldCount,
	}, nil
}

// validateSelect validates the SELECT clause.
func (v *Validator) validateSelect(vctx *validationContext, selects []*SelectExpression) error {
	if len(selects) == 0 {
		return NewValidationError(ErrCodeMissingRequiredClause, "SELECT clause is required")
	}

	for _, sel := range selects {
		if err := v.validateSelectExpression(vctx, sel); err != nil {
			return err
		}
	}

	// Check field count limit
	if err := v.limits.CheckSelectFields(vctx.fieldCount); err != nil {
		return err
	}

	return nil
}

// validateSelectExpression validates a single SELECT item.
func (v *Validator) validateSelectExpression(vctx *validationContext, sel *SelectExpression) error {
	if sel.Item == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty SELECT expression")
	}

	item := sel.Item

	switch {
	case item.Typeof != nil:
		return v.validateTypeof(vctx, item.Typeof)

	case item.Aggregate != nil:
		vctx.fieldCount++
		return v.validateAggregate(vctx, item.Aggregate)

	case item.Subquery != nil:
		return v.validateRelationshipSubquery(vctx, item.Subquery)

	case item.Expr != nil:
		vctx.fieldCount++
		return v.validateExpression(vctx, item.Expr)

	default:
		return NewValidationError(ErrCodeInvalidExpression, "invalid SELECT expression")
	}
}

// validateRelationshipSubquery validates a Parent-to-Child subquery.
func (v *Validator) validateRelationshipSubquery(vctx *validationContext, sub *RelationshipSubquery) error {
	if vctx.inSubquery {
		return NewValidationError(ErrCodeNestedSubqueryNotAllowed, "nested subqueries are not allowed")
	}

	// Find the relationship
	rel := vctx.rootObject.GetRelationship(sub.From)
	if rel == nil {
		return UnknownRelationshipErrorAt(vctx.rootObject.Name, sub.From, sub.Pos)
	}

	// Get child object metadata
	childObject, err := vctx.validator.metadata.GetObject(vctx.ctx, rel.ChildObject)
	if err != nil {
		return fmt.Errorf("failed to get child object metadata: %w", err)
	}
	if childObject == nil {
		return UnknownObjectError(rel.ChildObject)
	}

	// Check access to child object
	if err := vctx.validator.access.CanAccessObject(vctx.ctx, rel.ChildObject); err != nil {
		return err
	}

	// Create child validation context
	childCtx := &validationContext{
		ctx:          vctx.ctx,
		validator:    vctx.validator,
		rootObject:   childObject,
		resolvedRefs: make(map[string]*ResolvedRef),
		fieldCount:   0,
		lookupDepth:  0,
		inSubquery:   true,
		parentCtx:    vctx,
	}

	// Validate subquery SELECT
	for _, sel := range sub.Select {
		if err := v.validateSelectExpression(childCtx, sel); err != nil {
			return fmt.Errorf("in subquery %s: %w", sub.From, err)
		}
	}

	// Validate subquery WHERE
	if sub.Where != nil {
		if err := v.validateExpression(childCtx, sub.Where); err != nil {
			return fmt.Errorf("in subquery %s WHERE: %w", sub.From, err)
		}
	}

	// Validate subquery ORDER BY
	if err := v.validateOrderBy(childCtx, sub.OrderBy); err != nil {
		return fmt.Errorf("in subquery %s: %w", sub.From, err)
	}

	// Validate subquery LIMIT
	if sub.Limit != nil {
		if err := v.limits.CheckSubqueryRecords(*sub.Limit); err != nil {
			return err
		}
	}

	vctx.subqueries = append(vctx.subqueries, &ValidatedSubquery{
		Relationship: rel,
		ChildObject:  childObject,
		AST:          sub,
		FieldCount:   childCtx.fieldCount,
	})

	return nil
}

// validateAggregate validates an aggregate expression.
func (v *Validator) validateAggregate(vctx *validationContext, agg *AggregateExpression) error {
	if agg.Expression == nil {
		return NewValidationError(ErrCodeInvalidAggregation, "aggregate function requires an argument")
	}

	// Validate the inner expression
	if err := v.validateExpression(vctx, agg.Expression); err != nil {
		return fmt.Errorf("invalid aggregate argument: %w", err)
	}

	// For COUNT, any field type is allowed
	// For SUM, AVG - need numeric types
	// For MIN, MAX - need comparable types
	// Type checking will be done during type inference

	return nil
}

// validateExpression validates an expression.
func (v *Validator) validateExpression(vctx *validationContext, expr *Expression) error {
	if expr == nil || expr.Or == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty expression")
	}
	return v.validateOrExpr(vctx, expr.Or)
}

func (v *Validator) validateOrExpr(vctx *validationContext, or *OrExpr) error {
	for _, and := range or.And {
		if err := v.validateAndExpr(vctx, and); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateAndExpr(vctx *validationContext, and *AndExpr) error {
	for _, not := range and.Not {
		if err := v.validateNotExpr(vctx, not); err != nil {
			return err
		}
	}
	return nil
}

func (v *Validator) validateNotExpr(vctx *validationContext, not *NotExpr) error {
	return v.validateCompareExpr(vctx, not.Compare)
}

func (v *Validator) validateCompareExpr(vctx *validationContext, cmp *CompareExpr) error {
	if cmp.Left == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty comparison expression")
	}

	if err := v.validateInExpr(vctx, cmp.Left); err != nil {
		return err
	}

	if cmp.Right != nil {
		if err := v.validateInExpr(vctx, cmp.Right); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateInExpr(vctx *validationContext, in *InExpr) error {
	if err := v.validateLikeExpr(vctx, in.Left); err != nil {
		return err
	}

	if !in.In {
		return nil
	}

	// Handle WHERE subquery
	if in.Subquery != nil {
		return v.validateWhereSubquery(vctx, in.Subquery)
	}

	// Handle literal values
	for _, val := range in.Values {
		if err := v.validateValue(vctx, val); err != nil {
			return err
		}
	}

	return nil
}

// validateWhereSubquery validates a WHERE IN subquery (semi-join).
func (v *Validator) validateWhereSubquery(vctx *validationContext, sub *WhereSubquery) error {
	// Check for nested WHERE subqueries
	if vctx.inWhereSubquery {
		return NewValidationError(ErrCodeNestedSubqueryNotAllowed, "nested WHERE subqueries are not allowed")
	}

	// Get subquery object metadata
	subObject, err := v.metadata.GetObject(vctx.ctx, sub.From)
	if err != nil {
		return fmt.Errorf("failed to get subquery object metadata: %w", err)
	}
	if subObject == nil {
		return UnknownObjectError(sub.From)
	}

	// Check access to subquery object
	if err := v.access.CanAccessObject(vctx.ctx, sub.From); err != nil {
		return err
	}

	// The SELECT expression must be a simple field reference (no aggregates)
	if sub.Select == nil {
		return NewValidationError(ErrCodeInvalidExpression, "WHERE subquery SELECT clause is required")
	}

	// Extract the field from the SELECT expression
	// The expression should be a simple field reference
	selectedField, err := v.extractSingleFieldFromExpression(sub.Select, subObject, sub.Pos)
	if err != nil {
		return err
	}

	// Create subquery validation context
	subCtx := &validationContext{
		ctx:             vctx.ctx,
		validator:       v,
		rootObject:      subObject,
		resolvedRefs:    make(map[string]*ResolvedRef),
		fieldCount:      0,
		lookupDepth:     0,
		inSubquery:      true,
		inWhereSubquery: true,
		parentCtx:       vctx,
	}

	// Validate WHERE clause within subquery
	if sub.Where != nil {
		if err := v.validateExpression(subCtx, sub.Where); err != nil {
			return fmt.Errorf("in WHERE subquery WHERE clause: %w", err)
		}
	}

	// Validate LIMIT
	if sub.Limit != nil {
		if err := v.limits.CheckSubqueryRecords(*sub.Limit); err != nil {
			return err
		}
	}

	// Store validated WHERE subquery
	vctx.whereSubqueries = append(vctx.whereSubqueries, &ValidatedWhereSubquery{
		AST:          sub,
		Object:       subObject,
		Field:        selectedField,
		ResolvedRefs: subCtx.resolvedRefs,
	})

	return nil
}

// extractSingleFieldFromExpression extracts a single field from a SELECT expression in WHERE subquery.
// Returns error if the expression is not a simple field reference or contains aggregates.
func (v *Validator) extractSingleFieldFromExpression(expr *Expression, obj *ObjectMeta, pos lexer.Position) (*FieldMeta, error) {
	if expr == nil || expr.Or == nil || len(expr.Or.And) != 1 {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	and := expr.Or.And[0]
	if len(and.Not) != 1 {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	not := and.Not[0]
	if not.Not || not.Compare == nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	cmp := not.Compare
	if cmp.Operator != nil || cmp.Right != nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	in := cmp.Left
	if in == nil || in.In || in.Left == nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	like := in.Left
	if like.Like || like.Left == nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	is := like.Left
	if is.Is || is.Left == nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	add := is.Left
	if add.Left == nil || len(add.Right) > 0 {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	mul := add.Left
	if mul.Left == nil || len(mul.Right) > 0 {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	unary := mul.Left
	if unary.Operator != nil || unary.Primary == nil {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	primary := unary.Primary

	// Check for aggregate - not allowed in WHERE subquery
	if primary.Aggregate != nil {
		return nil, WhereSubqueryAggregateFieldError(pos)
	}

	// Must be a simple field reference
	if primary.Field == nil || len(primary.Field.Path) == 0 {
		return nil, WhereSubquerySingleFieldError(pos)
	}

	// Currently only support single field reference (no lookups)
	if len(primary.Field.Path) != 1 {
		return nil, NewValidationError(ErrCodeInvalidExpression, "lookup paths in WHERE subquery SELECT are not supported")
	}

	fieldName := primary.Field.Path[0]
	fieldMeta := obj.GetField(fieldName)
	if fieldMeta == nil {
		return nil, UnknownFieldErrorAt(obj.Name, fieldName, primary.Field.Pos)
	}

	return fieldMeta, nil
}

func (v *Validator) validateLikeExpr(vctx *validationContext, like *LikeExpr) error {
	if err := v.validateIsExpr(vctx, like.Left); err != nil {
		return err
	}

	if like.Pattern != nil {
		if err := v.validateValue(vctx, like.Pattern); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateIsExpr(vctx *validationContext, is *IsExpr) error {
	return v.validateAddExpr(vctx, is.Left)
}

func (v *Validator) validateAddExpr(vctx *validationContext, add *AddExpr) error {
	if err := v.validateMulExpr(vctx, add.Left); err != nil {
		return err
	}

	for _, op := range add.Right {
		if err := v.validateMulExpr(vctx, op.Right); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateMulExpr(vctx *validationContext, mul *MulExpr) error {
	if err := v.validateUnaryExpr(vctx, mul.Left); err != nil {
		return err
	}

	for _, op := range mul.Right {
		if err := v.validateUnaryExpr(vctx, op.Right); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateUnaryExpr(vctx *validationContext, unary *UnaryExpr) error {
	return v.validatePrimary(vctx, unary.Primary)
}

func (v *Validator) validatePrimary(vctx *validationContext, primary *Primary) error {
	if primary == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty primary expression")
	}

	switch {
	case primary.Subexpression != nil:
		return v.validateExpression(vctx, primary.Subexpression)

	case primary.Aggregate != nil:
		return v.validateAggregate(vctx, primary.Aggregate)

	case primary.FuncCall != nil:
		return v.validateFuncCall(vctx, primary.FuncCall)

	case primary.Const != nil:
		return v.validateConst(vctx, primary.Const)

	case primary.Field != nil:
		return v.validateField(vctx, primary.Field, true)

	default:
		return NewValidationError(ErrCodeInvalidExpression, "invalid primary expression")
	}
}

func (v *Validator) validateValue(vctx *validationContext, val *Value) error {
	if val == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty value")
	}

	if val.Const != nil {
		return v.validateConst(vctx, val.Const)
	}

	if val.Field != nil {
		return v.validateField(vctx, val.Field, true)
	}

	return NewValidationError(ErrCodeInvalidExpression, "invalid value")
}

func (v *Validator) validateFuncCall(vctx *validationContext, fn *FuncCall) error {
	// Validate argument count
	argCount := len(fn.Args)
	minArgs := fn.Name.MinArgs()
	maxArgs := fn.Name.MaxArgs()

	if argCount < minArgs {
		return NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("%s requires at least %d argument(s), got %d", fn.Name, minArgs, argCount))
	}

	if maxArgs >= 0 && argCount > maxArgs {
		return NewValidationError(ErrCodeInvalidExpression,
			fmt.Sprintf("%s accepts at most %d argument(s), got %d", fn.Name, maxArgs, argCount))
	}

	// Validate each argument
	for _, arg := range fn.Args {
		if err := v.validateExpression(vctx, arg); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateConst(vctx *validationContext, c *Const) error {
	// Constants are always valid, just ensure type can be inferred
	c.GetFieldType()
	return nil
}

// validateField validates a field reference and resolves it through lookups.
func (v *Validator) validateField(vctx *validationContext, field *Field, checkFilterable bool) error {
	if field == nil || len(field.Path) == 0 {
		return NewValidationError(ErrCodeInvalidExpression, "empty field reference")
	}

	pathKey := strings.Join(field.Path, ".")

	// Check if already resolved
	if _, ok := vctx.resolvedRefs[pathKey]; ok {
		return nil
	}

	// Resolve the field path
	resolved, err := v.resolveFieldPath(vctx, field.Path, checkFilterable, field.Pos)
	if err != nil {
		return err
	}

	vctx.resolvedRefs[pathKey] = resolved
	field.FieldType = resolved.Field.Type

	return nil
}

// resolveFieldPath resolves a field path through lookups.
// Example: Account.Owner.Name resolves through Contact -> Account -> User
func (v *Validator) resolveFieldPath(vctx *validationContext, path []string, checkFilterable bool, pos lexer.Position) (*ResolvedRef, error) {
	if len(path) == 0 {
		return nil, NewValidationError(ErrCodeInvalidExpression, "empty field path")
	}

	currentObject := vctx.rootObject
	var joins []*Join
	var lookupPath []string

	// Process path except last element (which is the field)
	for i := 0; i < len(path)-1; i++ {
		lookupName := path[i]

		// Check lookup depth
		if vctx.lookupDepth+i+1 > v.limits.MaxLookupDepth && v.limits.MaxLookupDepth > 0 {
			return nil, NewLimitError(LimitTypeMaxLookupDepth, v.limits.MaxLookupDepth, vctx.lookupDepth+i+1)
		}

		// Find lookup
		lookup := currentObject.GetLookup(lookupName)
		if lookup == nil {
			return nil, UnknownLookupErrorAt(currentObject.Name, lookupName, pos)
		}

		// Get target object
		targetObject, err := v.metadata.GetObject(vctx.ctx, lookup.TargetObject)
		if err != nil {
			return nil, fmt.Errorf("failed to get lookup target: %w", err)
		}
		if targetObject == nil {
			return nil, UnknownObjectError(lookup.TargetObject)
		}

		// Check access to target object
		if err := v.access.CanAccessObject(vctx.ctx, lookup.TargetObject); err != nil {
			return nil, err
		}

		// Add join
		joins = append(joins, &Join{
			FromObject: currentObject,
			FromField:  lookup.Field,
			ToObject:   targetObject,
			ToField:    lookup.TargetField,
			Alias:      strings.Join(path[:i+1], "_"),
		})

		lookupPath = append(lookupPath, currentObject.Name)
		currentObject = targetObject
	}

	// Resolve the final field
	fieldName := path[len(path)-1]
	fieldMeta := currentObject.GetField(fieldName)
	if fieldMeta == nil {
		return nil, UnknownFieldErrorAt(currentObject.Name, fieldName, pos)
	}

	// Check field access
	if err := v.access.CanAccessField(vctx.ctx, currentObject.Name, fieldName); err != nil {
		return nil, err
	}

	// Check if field is filterable (for WHERE clause)
	if checkFilterable && !fieldMeta.Filterable {
		return nil, FieldNotFilterableError(currentObject.Name, fieldName)
	}

	return &ResolvedRef{
		Path:       path,
		Object:     currentObject,
		Field:      fieldMeta,
		Joins:      joins,
		LookupPath: lookupPath,
	}, nil
}

// validateGroupBy validates the GROUP BY clause.
func (v *Validator) validateGroupBy(vctx *validationContext, groups []*GroupClause) error {
	for _, group := range groups {
		if len(group.Field) == 0 {
			return NewValidationError(ErrCodeInvalidExpression, "empty GROUP BY field")
		}

		// Resolve the field
		pathKey := strings.Join(group.Field, ".")
		if _, ok := vctx.resolvedRefs[pathKey]; !ok {
			resolved, err := v.resolveFieldPath(vctx, group.Field, false, group.Pos)
			if err != nil {
				return fmt.Errorf("invalid GROUP BY field: %w", err)
			}

			// Check if field is groupable
			if !resolved.Field.Groupable {
				return FieldNotGroupableError(resolved.Object.Name, resolved.Field.Name)
			}

			vctx.resolvedRefs[pathKey] = resolved
		}
	}

	return nil
}

// validateOrderBy validates the ORDER BY clause.
func (v *Validator) validateOrderBy(vctx *validationContext, orders []*OrderClause) error {
	for _, order := range orders {
		if order.OrderItem == nil {
			return NewValidationError(ErrCodeInvalidExpression, "empty ORDER BY item")
		}

		item := order.OrderItem

		if item.Aggregate != nil {
			// Validate aggregate in ORDER BY
			if err := v.validateAggregate(vctx, item.Aggregate); err != nil {
				return fmt.Errorf("invalid ORDER BY aggregate: %w", err)
			}
		} else if len(item.Field) > 0 {
			// Validate field in ORDER BY
			pathKey := strings.Join(item.Field, ".")
			if _, ok := vctx.resolvedRefs[pathKey]; !ok {
				resolved, err := v.resolveFieldPath(vctx, item.Field, false, item.Pos)
				if err != nil {
					return fmt.Errorf("invalid ORDER BY field: %w", err)
				}

				// Check if field is sortable
				if !resolved.Field.Sortable {
					return FieldNotSortableError(resolved.Object.Name, resolved.Field.Name)
				}

				vctx.resolvedRefs[pathKey] = resolved
			}
		} else {
			return NewValidationError(ErrCodeInvalidExpression, "ORDER BY requires field or aggregate")
		}
	}

	return nil
}

// validateTypeof validates a TYPEOF expression.
func (v *Validator) validateTypeof(vctx *validationContext, typeof *TypeofExpression) error {
	if typeof == nil {
		return NewValidationError(ErrCodeInvalidExpression, "empty TYPEOF expression")
	}

	// Validate that TYPEOF is not used in subqueries
	if vctx.inSubquery {
		return NewValidationError(ErrCodeInvalidExpression, "TYPEOF is not allowed in subqueries")
	}

	// Check that the polymorphic field exists on the root object
	// For now, we just check that it's a valid field name
	fieldMeta := vctx.rootObject.GetField(typeof.Field)
	if fieldMeta == nil {
		return UnknownFieldErrorAt(vctx.rootObject.Name, typeof.Field, typeof.Pos)
	}

	// Check field access
	if err := v.access.CanAccessField(vctx.ctx, vctx.rootObject.Name, typeof.Field); err != nil {
		return err
	}

	validated := &ValidatedTypeof{
		AST:         typeof,
		Field:       typeof.Field,
		WhenClauses: make([]*ValidatedTypeofWhen, 0, len(typeof.WhenClauses)),
	}

	// Validate each WHEN clause
	for _, when := range typeof.WhenClauses {
		validatedWhen, err := v.validateTypeofWhen(vctx, when)
		if err != nil {
			return err
		}
		validated.WhenClauses = append(validated.WhenClauses, validatedWhen)
		// Count fields from each WHEN clause
		vctx.fieldCount += len(when.Fields)
	}

	// Validate ELSE fields if present
	if len(typeof.ElseFields) > 0 {
		// ELSE fields need to exist on at least one of the possible object types
		// For simplicity, we just count them for now
		vctx.fieldCount += len(typeof.ElseFields)
	}

	vctx.typeofExpressions = append(vctx.typeofExpressions, validated)

	return nil
}

// validateTypeofWhen validates a single WHEN clause in TYPEOF.
func (v *Validator) validateTypeofWhen(vctx *validationContext, when *WhenClause) (*ValidatedTypeofWhen, error) {
	// Get the target object metadata
	targetObject, err := v.metadata.GetObject(vctx.ctx, when.ObjectType)
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata for TYPEOF WHEN: %w", err)
	}
	if targetObject == nil {
		return nil, UnknownObjectError(when.ObjectType)
	}

	// Check access to target object
	if err := v.access.CanAccessObject(vctx.ctx, when.ObjectType); err != nil {
		return nil, err
	}

	validatedWhen := &ValidatedTypeofWhen{
		ObjectType: when.ObjectType,
		Object:     targetObject,
		Fields:     make([]*FieldMeta, 0, len(when.Fields)),
	}

	// Validate each field in the WHEN clause
	for _, fieldName := range when.Fields {
		fieldMeta := targetObject.GetField(fieldName)
		if fieldMeta == nil {
			return nil, UnknownFieldErrorAt(when.ObjectType, fieldName, when.Pos)
		}

		// Check field access
		if err := v.access.CanAccessField(vctx.ctx, when.ObjectType, fieldName); err != nil {
			return nil, err
		}

		validatedWhen.Fields = append(validatedWhen.Fields, fieldMeta)
	}

	return validatedWhen, nil
}
