package engine

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// quoteIdent quotes a single SQL identifier using pgx.
func quoteIdent(name string) string {
	return pgx.Identifier{name}.Sanitize()
}

// qualifiedColumn returns a quoted alias.column expression.
func qualifiedColumn(alias, column string) string {
	return alias + "." + quoteIdent(column)
}

// CompiledQuery represents a compiled SOQL query ready for execution.
type CompiledQuery struct {
	SQL        string       // SQL with placeholders ($1, $2...)
	Params     []any        // Static parameters
	DateParams []*DateParam // Date literal parameters for runtime resolution
	Shape      *ResultShape // Structure of expected result

	// Pagination contains keyset pagination metadata.
	// Used by executor to handle cursor-based pagination.
	Pagination *PaginationInfo

	// Dependencies contains the list of object API names that this query depends on.
	// Used for targeted cache invalidation when metadata changes.
	Dependencies []string

	// ForUpdate indicates that the query uses FOR UPDATE locking.
	// When true, the query should be executed within a transaction.
	ForUpdate bool

	// WithSecurityEnforced indicates that FLS/OLS security checks should be enforced.
	// When true, the query should fail if the user lacks access to any field or object,
	// rather than silently filtering data.
	WithSecurityEnforced bool
}

// DateParam represents a date literal parameter that needs runtime resolution.
type DateParam struct {
	ParamIndex int                 // Index in Params slice where the resolved value will be
	Static     *StaticDateLiteral  // Static date literal (TODAY, THIS_WEEK, etc.)
	Dynamic    *DynamicDateLiteral // Dynamic date literal (LAST_N_DAYS:30, etc.)
	IsRange    bool                // Whether this is a range comparison (BETWEEN)
	EndIndex   int                 // End index for range comparisons
}

// ResultShape describes the structure of the query result.
type ResultShape struct {
	Object        string               // SOQL object name
	Table         string               // SQL table name
	Fields        []*FieldShape        // Fields in order
	Relationships []*RelationshipShape // Child relationships (subqueries)
}

// FieldShape describes a single field in the result.
type FieldShape struct {
	Name   string    // SOQL field name or alias
	Column string    // SQL column expression
	Type   FieldType // Field type
	Alias  string    // SQL alias used in query
}

// RelationshipShape describes a child relationship subquery result.
type RelationshipShape struct {
	Name  string       // Relationship name
	Shape *ResultShape // Nested result shape
}

// Compiler compiles validated SOQL queries to SQL.
type Compiler struct {
	limits *Limits
}

// NewCompiler creates a new Compiler.
func NewCompiler(limits *Limits) *Compiler {
	if limits == nil {
		limits = &DefaultLimits
	}
	return &Compiler{limits: limits}
}

// compileContext holds state during compilation.
type compileContext struct {
	compiler        *Compiler
	validated       *ValidatedQuery
	params          []any
	dateParams      []*DateParam
	paramCount      int
	joinAliases     map[string]string // join path -> alias
	joinSQL         []string          // JOIN clauses
	shape           *ResultShape
	mainAlias       string
	whereSubqueries []*ValidatedWhereSubquery // WHERE subqueries from validation

	// Keyset pagination fields
	keysetFields []*KeysetField // ORDER BY fields for keyset pagination
}

func newCompileContext(c *Compiler, v *ValidatedQuery) *compileContext {
	return &compileContext{
		compiler:        c,
		validated:       v,
		params:          make([]any, 0),
		dateParams:      make([]*DateParam, 0),
		paramCount:      0,
		joinAliases:     make(map[string]string),
		joinSQL:         make([]string, 0),
		shape:           &ResultShape{},
		mainAlias:       "t0",
		whereSubqueries: v.WhereSubqueries,
		keysetFields:    make([]*KeysetField, 0),
	}
}

// Compile compiles a validated SOQL query to SQL.
func (c *Compiler) Compile(validated *ValidatedQuery) (*CompiledQuery, error) {
	ctx := newCompileContext(c, validated)

	// Initialize result shape
	ctx.shape.Object = validated.RootObject.Name
	ctx.shape.Table = validated.RootObject.QualifiedTableName()

	// Build SELECT clause
	selectSQL, err := c.compileSelect(ctx, validated.AST.Select)
	if err != nil {
		return nil, fmt.Errorf("failed to compile SELECT: %w", err)
	}

	// Build FROM clause
	fromSQL := fmt.Sprintf("%s AS %s", validated.RootObject.QualifiedTableName(), ctx.mainAlias)

	// Build JOIN clauses from resolved references
	if err := c.buildJoins(ctx); err != nil {
		return nil, fmt.Errorf("failed to build JOINs: %w", err)
	}

	// Build WHERE clause
	var whereSQL string
	if validated.AST.Where != nil {
		whereSQL, err = c.compileExpression(ctx, validated.AST.Where)
		if err != nil {
			return nil, fmt.Errorf("failed to compile WHERE: %w", err)
		}
	}

	// Build GROUP BY clause
	var groupBySQL string
	if len(validated.AST.GroupBy) > 0 {
		groupBySQL, err = c.compileGroupBy(ctx, validated.AST.GroupBy)
		if err != nil {
			return nil, fmt.Errorf("failed to compile GROUP BY: %w", err)
		}
	}

	// Build HAVING clause
	var havingSQL string
	if validated.AST.Having != nil {
		havingSQL, err = c.compileExpression(ctx, validated.AST.Having)
		if err != nil {
			return nil, fmt.Errorf("failed to compile HAVING: %w", err)
		}
	}

	// Build ORDER BY clause
	hasExplicitOrderBy := len(validated.AST.OrderBy) > 0
	var orderBySQL string
	if hasExplicitOrderBy {
		orderBySQL, err = c.compileOrderBy(ctx, validated.AST.OrderBy)
		if err != nil {
			return nil, fmt.Errorf("failed to compile ORDER BY: %w", err)
		}
	}

	// For keyset pagination, ensure we have a tie-breaker
	// Add id as tie-breaker if not already present
	tieBreaker := c.ensureTieBreaker(ctx, validated.RootObject)
	if tieBreaker != "" {
		if orderBySQL != "" {
			orderBySQL += ", " + tieBreaker
		} else {
			// Default order: use tie-breaker as the only ORDER BY
			orderBySQL = tieBreaker
		}
	}

	// Add keyset fields to SELECT for cursor building
	// These fields are needed even if not explicitly selected
	keysetSelectSQL := c.ensureKeysetFieldsInSelect(ctx)
	if keysetSelectSQL != "" {
		selectSQL += ", " + keysetSelectSQL
	}

	// Build LIMIT (OFFSET is not used with keyset pagination)
	limit := c.limits.EffectiveLimit(validated.AST.Limit)
	var limitSQL string
	if limit > 0 {
		limitSQL = fmt.Sprintf("LIMIT %d", limit)
	}

	// Assemble final SQL
	var sql strings.Builder
	sql.WriteString("SELECT ")
	sql.WriteString(selectSQL)
	sql.WriteString("\nFROM ")
	sql.WriteString(fromSQL)

	for _, join := range ctx.joinSQL {
		sql.WriteString("\n")
		sql.WriteString(join)
	}

	if whereSQL != "" {
		sql.WriteString("\nWHERE ")
		sql.WriteString(whereSQL)
	}

	if groupBySQL != "" {
		sql.WriteString("\nGROUP BY ")
		sql.WriteString(groupBySQL)
	}

	if havingSQL != "" {
		sql.WriteString("\nHAVING ")
		sql.WriteString(havingSQL)
	}

	if orderBySQL != "" {
		sql.WriteString("\nORDER BY ")
		sql.WriteString(orderBySQL)
	}

	if limitSQL != "" {
		sql.WriteString("\n")
		sql.WriteString(limitSQL)
	}

	// Add FOR UPDATE clause if requested
	if validated.AST.ForUpdate {
		sql.WriteString("\nFOR UPDATE")
	}

	// Build pagination info for keyset cursor
	pagination := c.buildPaginationInfo(ctx, validated.RootObject.Name, hasExplicitOrderBy, limit)

	return &CompiledQuery{
		SQL:                  sql.String(),
		Params:               ctx.params,
		DateParams:           ctx.dateParams,
		Shape:                ctx.shape,
		Pagination:           pagination,
		Dependencies:         c.collectDependencies(validated),
		ForUpdate:            validated.AST.ForUpdate,
		WithSecurityEnforced: validated.AST.WithSecurityEnforced,
	}, nil
}

// collectDependencies extracts all object API names that the query depends on.
// This includes the root object, objects from lookups (joins), objects from subqueries,
// and objects from WHERE subqueries.
func (c *Compiler) collectDependencies(validated *ValidatedQuery) []string {
	deps := make(map[string]struct{})

	// Root object
	if validated.RootObject != nil {
		deps[validated.RootObject.Name] = struct{}{}
	}

	// Objects from lookups (joins)
	for _, ref := range validated.ResolvedRefs {
		for _, join := range ref.Joins {
			if join.ToObject != nil {
				deps[join.ToObject.Name] = struct{}{}
			}
		}
	}

	// Objects from relationship subqueries (SELECT)
	for _, sub := range validated.Subqueries {
		if sub.ChildObject != nil {
			deps[sub.ChildObject.Name] = struct{}{}
		}
	}

	// Objects from WHERE subqueries
	for _, sub := range validated.WhereSubqueries {
		if sub.Object != nil {
			deps[sub.Object.Name] = struct{}{}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(deps))
	for dep := range deps {
		result = append(result, dep)
	}
	return result
}

// ensureTieBreaker adds a tie-breaker to ORDER BY if not already present.
// Returns the tie-breaker SQL expression or empty string if already present.
func (c *Compiler) ensureTieBreaker(ctx *compileContext, rootObject *ObjectMeta) string {
	tieBreaker := DefaultTieBreaker

	// Check if tie-breaker is already in keyset fields
	for _, kf := range ctx.keysetFields {
		if kf.SQLColumn == tieBreaker {
			return "" // Already present
		}
	}

	// Add tie-breaker with same direction as first field (or DESC if no fields)
	direction := "desc"
	if len(ctx.keysetFields) > 0 {
		direction = ctx.keysetFields[0].Direction
	}

	// Add to keyset fields
	ctx.keysetFields = append(ctx.keysetFields, &KeysetField{
		SOQLName:   "Id",
		SQLColumn:  tieBreaker,
		TableAlias: ctx.mainAlias,
		Direction:  direction,
	})

	// Return SQL expression
	expr := qualifiedColumn(ctx.mainAlias, tieBreaker)
	if direction == "desc" {
		return expr + " DESC"
	}
	return expr
}

// ensureKeysetFieldsInSelect adds keyset pagination fields to SELECT if not already present.
// These fields are needed for cursor building even if not explicitly selected by user.
func (c *Compiler) ensureKeysetFieldsInSelect(ctx *compileContext) string {
	if len(ctx.keysetFields) == 0 {
		return ""
	}

	// Build a set of already selected column names (by SQL column name)
	selectedCols := make(map[string]bool)
	for _, f := range ctx.shape.Fields {
		selectedCols[f.Column] = true
		// Also check the alias in case it matches the SOQL name
		if f.Alias != "" {
			selectedCols[f.Alias] = true
		}
		if f.Name != "" {
			selectedCols[f.Name] = true
		}
	}

	var parts []string
	for _, kf := range ctx.keysetFields {
		// Check if this keyset field is already in SELECT
		fullColumn := qualifiedColumn(kf.TableAlias, kf.SQLColumn)
		if selectedCols[kf.SQLColumn] || selectedCols[fullColumn] || selectedCols[kf.SOQLName] {
			continue
		}

		// Add to SELECT with SOQL name as alias
		part := fullColumn + " AS " + kf.SOQLName
		parts = append(parts, part)

		// Add to shape for result parsing
		ctx.shape.Fields = append(ctx.shape.Fields, &FieldShape{
			Name:   kf.SOQLName,
			Column: kf.SQLColumn,
			Type:   FieldTypeID, // Keyset fields are typically IDs or sortable types
			Alias:  kf.SOQLName,
		})
	}

	return strings.Join(parts, ", ")
}

// buildPaginationInfo creates PaginationInfo from keyset fields.
func (c *Compiler) buildPaginationInfo(ctx *compileContext, objectName string, hasOrderBy bool, pageSize int) *PaginationInfo {
	if len(ctx.keysetFields) == 0 {
		return nil
	}

	sortKeys := make([]struct {
		Field string `json:"f"`
		Dir   string `json:"d"`
	}, len(ctx.keysetFields))

	sortKeySOQL := make([]string, len(ctx.keysetFields))

	for i, kf := range ctx.keysetFields {
		sortKeys[i].Field = kf.SQLColumn
		sortKeys[i].Dir = kf.Direction
		sortKeySOQL[i] = kf.SOQLName
	}

	// Convert to keyset.SortKeys format
	ksKeys := make([]struct {
		Field string `json:"f"`
		Dir   string `json:"d"`
	}, len(sortKeys))
	copy(ksKeys, sortKeys)

	return &PaginationInfo{
		SortKeys:    convertToKeysetSortKeys(ctx.keysetFields),
		SortKeySOQL: sortKeySOQL,
		TieBreaker:  DefaultTieBreaker,
		PageSize:    pageSize,
		HasOrderBy:  hasOrderBy,
		Object:      objectName,
	}
}

// convertToKeysetSortKeys converts KeysetField slice to SortKeys.
func convertToKeysetSortKeys(fields []*KeysetField) SortKeys {
	keys := make(SortKeys, len(fields))
	for i, f := range fields {
		keys[i] = SortKey{
			Field: f.SQLColumn,
			Dir:   SortDirection(f.Direction),
		}
	}
	return keys
}

// compileSelect compiles the SELECT clause.
func (c *Compiler) compileSelect(ctx *compileContext, selects []*SelectExpression) (string, error) {
	var parts []string

	for i, sel := range selects {
		part, err := c.compileSelectExpression(ctx, sel, i)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, ", "), nil
}

// compileSelectExpression compiles a single SELECT item.
func (c *Compiler) compileSelectExpression(ctx *compileContext, sel *SelectExpression, index int) (string, error) {
	if sel.Item == nil {
		return "", fmt.Errorf("empty SELECT expression")
	}

	item := sel.Item
	var expr string
	var fieldType FieldType
	var err error

	switch {
	case item.Typeof != nil:
		expr, err = c.compileTypeof(ctx, item.Typeof)
		fieldType = FieldTypeObject // TYPEOF returns JSON object

	case item.Aggregate != nil:
		expr, err = c.compileAggregate(ctx, item.Aggregate)
		fieldType = item.Aggregate.FieldType

	case item.Subquery != nil:
		expr, err = c.compileSubquery(ctx, item.Subquery)
		fieldType = FieldTypeArray

	case item.Expr != nil:
		expr, err = c.compileExpression(ctx, item.Expr)
		fieldType = item.Expr.FieldType
	}

	if err != nil {
		return "", err
	}

	// Determine alias
	alias := sel.Alias
	if alias == nil {
		// Extract natural alias from expression
		defaultAlias := extractNaturalAlias(sel)
		alias = &defaultAlias
	}

	// Add to shape
	ctx.shape.Fields = append(ctx.shape.Fields, &FieldShape{
		Name:   *alias,
		Column: expr,
		Type:   fieldType,
		Alias:  *alias,
	})

	return fmt.Sprintf("%s AS %s", expr, *alias), nil
}

// extractNaturalAlias extracts a meaningful alias from a SelectExpression.
// For simple fields: uses field name (e.g., "name" → "name")
// For lookup fields: joins path with underscore (e.g., "Account.Name" → "Account_Name")
// For aggregates: uses function + field (e.g., "COUNT(Id)" → "COUNT_Id")
// For subqueries: uses relationship name (e.g., "(SELECT ... FROM Contacts)" → "Contacts")
func extractNaturalAlias(sel *SelectExpression) string {
	if sel.Item == nil {
		return "expr"
	}

	// TYPEOF → field name
	if sel.Item.Typeof != nil {
		return sel.Item.Typeof.Field
	}

	// Subquery → relationship name
	if sel.Item.Subquery != nil {
		return sel.Item.Subquery.From
	}

	// Aggregate → FUNC_field
	if sel.Item.Aggregate != nil {
		fieldName := extractFieldNameFromExpr(sel.Item.Aggregate.Expression)
		if fieldName == "" {
			fieldName = "expr"
		}
		return sel.Item.Aggregate.Function.String() + "_" + fieldName
	}

	// Expression → field path
	if sel.Item.Expr != nil {
		return extractFieldNameFromExpr(sel.Item.Expr)
	}

	return "expr"
}

// extractFieldNameFromExpr extracts field name/path from an Expression.
func extractFieldNameFromExpr(expr *Expression) string {
	if expr == nil || expr.Or == nil {
		return ""
	}
	return extractFieldNameFromOr(expr.Or)
}

func extractFieldNameFromOr(or *OrExpr) string {
	if or == nil || len(or.And) == 0 {
		return ""
	}
	return extractFieldNameFromAnd(or.And[0])
}

func extractFieldNameFromAnd(and *AndExpr) string {
	if and == nil || len(and.Not) == 0 {
		return ""
	}
	return extractFieldNameFromNot(and.Not[0])
}

func extractFieldNameFromNot(not *NotExpr) string {
	if not == nil || not.Compare == nil {
		return ""
	}
	return extractFieldNameFromCompare(not.Compare)
}

func extractFieldNameFromCompare(cmp *CompareExpr) string {
	if cmp == nil || cmp.Left == nil {
		return ""
	}
	return extractFieldNameFromIn(cmp.Left)
}

func extractFieldNameFromIn(in *InExpr) string {
	if in == nil || in.Left == nil {
		return ""
	}
	return extractFieldNameFromLike(in.Left)
}

func extractFieldNameFromLike(like *LikeExpr) string {
	if like == nil || like.Left == nil {
		return ""
	}
	return extractFieldNameFromIs(like.Left)
}

func extractFieldNameFromIs(is *IsExpr) string {
	if is == nil || is.Left == nil {
		return ""
	}
	return extractFieldNameFromAdd(is.Left)
}

func extractFieldNameFromAdd(add *AddExpr) string {
	if add == nil || add.Left == nil {
		return ""
	}
	return extractFieldNameFromMul(add.Left)
}

func extractFieldNameFromMul(mul *MulExpr) string {
	if mul == nil || mul.Left == nil {
		return ""
	}
	return extractFieldNameFromUnary(mul.Left)
}

func extractFieldNameFromUnary(unary *UnaryExpr) string {
	if unary == nil || unary.Primary == nil {
		return ""
	}
	return extractFieldNameFromPrimary(unary.Primary)
}

func extractFieldNameFromPrimary(primary *Primary) string {
	if primary == nil {
		return ""
	}

	// Field reference
	if primary.Field != nil && len(primary.Field.Path) > 0 {
		return strings.Join(primary.Field.Path, "_")
	}

	// Nested expression
	if primary.Subexpression != nil {
		return extractFieldNameFromExpr(primary.Subexpression)
	}

	// Function call
	if primary.FuncCall != nil {
		return primary.FuncCall.Name.String()
	}

	return ""
}

// compileAggregate compiles an aggregate expression.
func (c *Compiler) compileAggregate(ctx *compileContext, agg *AggregateExpression) (string, error) {
	inner, err := c.compileExpression(ctx, agg.Expression)
	if err != nil {
		return "", err
	}

	switch agg.Function {
	case AggregateCount:
		return fmt.Sprintf("COUNT(%s)", inner), nil
	case AggregateCountDistinct:
		return fmt.Sprintf("COUNT(DISTINCT %s)", inner), nil
	case AggregateSum:
		return fmt.Sprintf("SUM(%s)", inner), nil
	case AggregateAvg:
		return fmt.Sprintf("AVG(%s)", inner), nil
	case AggregateMin:
		return fmt.Sprintf("MIN(%s)", inner), nil
	case AggregateMax:
		return fmt.Sprintf("MAX(%s)", inner), nil
	default:
		return fmt.Sprintf("COUNT(%s)", inner), nil
	}
}

// compileTypeof compiles a TYPEOF expression to SQL CASE expression with JSON_BUILD_OBJECT.
// TYPEOF allows conditional field selection based on the polymorphic field's actual type.
// Example SOQL:
//
//	TYPEOF What
//	    WHEN Account THEN Name, Industry
//	    WHEN Opportunity THEN Name, StageName
//	    ELSE Name
//	END
//
// Generates SQL like:
//
//	CASE
//	    WHEN t0.what_type = 'Account' THEN JSON_BUILD_OBJECT('type', 'Account', 'Name', t_acct.name, 'Industry', t_acct.industry)
//	    WHEN t0.what_type = 'Opportunity' THEN JSON_BUILD_OBJECT('type', 'Opportunity', 'Name', t_opp.name, 'StageName', t_opp.stage_name)
//	    ELSE JSON_BUILD_OBJECT('Name', ...)
//	END
func (c *Compiler) compileTypeof(ctx *compileContext, typeof *TypeofExpression) (string, error) {
	// Find the validated TYPEOF expression
	var validatedTypeof *ValidatedTypeof
	for _, vt := range ctx.validated.TypeofExpressions {
		if vt.AST == typeof {
			validatedTypeof = vt
			break
		}
	}
	if validatedTypeof == nil {
		return "", fmt.Errorf("TYPEOF expression not validated: %s", typeof.Field)
	}

	// Build CASE expression
	var caseParts []string
	caseParts = append(caseParts, "CASE")

	// Get the polymorphic field info
	// We assume the polymorphic field has a corresponding _type column
	// e.g., what_id and what_type for a polymorphic "What" field
	polyField := ctx.validated.RootObject.GetField(typeof.Field)
	if polyField == nil {
		return "", fmt.Errorf("polymorphic field not found: %s", typeof.Field)
	}

	// Type column name - derive from Column by replacing _id with _type
	// e.g., "what_id" -> "what_type", or if no _id suffix, append _type
	baseColumn := polyField.Column
	if strings.HasSuffix(baseColumn, "_id") {
		baseColumn = strings.TrimSuffix(baseColumn, "_id")
	}
	typeColumn := qualifiedColumn(ctx.mainAlias, baseColumn+"_type")

	// Process each WHEN clause
	for _, when := range validatedTypeof.WhenClauses {
		// Generate JOIN for this object type if needed
		joinAlias := c.getOrCreateTypeofJoin(ctx, typeof.Field, when.ObjectType, when.Object)

		// Build JSON object for this type's fields
		var jsonParts []string
		jsonParts = append(jsonParts, fmt.Sprintf("'type', '%s'", when.ObjectType))

		for _, field := range when.Fields {
			fieldColumn := qualifiedColumn(joinAlias, field.Column)
			jsonParts = append(jsonParts, fmt.Sprintf("'%s', %s", field.Name, fieldColumn))
		}

		whenSQL := fmt.Sprintf("WHEN %s = '%s' THEN JSON_BUILD_OBJECT(%s)",
			typeColumn, when.ObjectType, strings.Join(jsonParts, ", "))
		caseParts = append(caseParts, whenSQL)
	}

	// Handle ELSE clause
	if len(typeof.ElseFields) > 0 {
		var elseParts []string
		elseParts = append(elseParts, "'type', 'Unknown'")
		for _, fieldName := range typeof.ElseFields {
			// For ELSE, we try to get the field from the main object
			fieldMeta := ctx.validated.RootObject.GetField(fieldName)
			if fieldMeta != nil {
				elseParts = append(elseParts, fmt.Sprintf("'%s', %s", fieldName, qualifiedColumn(ctx.mainAlias, fieldMeta.Column)))
			}
		}
		caseParts = append(caseParts, fmt.Sprintf("ELSE JSON_BUILD_OBJECT(%s)", strings.Join(elseParts, ", ")))
	} else {
		// Default ELSE returns NULL
		caseParts = append(caseParts, "ELSE NULL")
	}

	caseParts = append(caseParts, "END")

	return strings.Join(caseParts, " "), nil
}

// getOrCreateTypeofJoin creates or returns existing join for a TYPEOF WHEN clause.
func (c *Compiler) getOrCreateTypeofJoin(ctx *compileContext, polyFieldName, objectType string, obj *ObjectMeta) string {
	// Create a unique join path for this typeof/object combination
	joinPath := fmt.Sprintf("typeof_%s_%s", polyFieldName, objectType)

	if alias, ok := ctx.joinAliases[joinPath]; ok {
		return alias
	}

	// Get the polymorphic field's column name
	polyField := ctx.validated.RootObject.GetField(polyFieldName)
	if polyField == nil {
		// Fallback to lowercase field name
		return ""
	}

	// Create new join alias
	alias := fmt.Sprintf("t%d", len(ctx.joinAliases)+1)
	ctx.joinAliases[joinPath] = alias

	// Derive id and type columns from the polymorphic field's Column
	// e.g., "what_id" -> idColumn="what_id", typeColumn="what_type"
	baseColumn := polyField.Column
	idColumn := baseColumn
	if strings.HasSuffix(baseColumn, "_id") {
		baseColumn = strings.TrimSuffix(baseColumn, "_id")
	}
	typeColumn := baseColumn + "_type"

	// Polymorphic join: LEFT JOIN on both id and type match
	joinSQL := fmt.Sprintf("LEFT JOIN %s AS %s ON %s = %s AND %s = '%s'",
		obj.QualifiedTableName(), alias,
		qualifiedColumn(ctx.mainAlias, idColumn), qualifiedColumn(alias, "id"),
		qualifiedColumn(ctx.mainAlias, typeColumn), objectType)

	ctx.joinSQL = append(ctx.joinSQL, joinSQL)

	return alias
}

// compileSubquery compiles a Parent-to-Child relationship subquery using JSON_AGG.
func (c *Compiler) compileSubquery(ctx *compileContext, sub *RelationshipSubquery) (string, error) {
	// Find the validated subquery
	var validatedSub *ValidatedSubquery
	for _, vs := range ctx.validated.Subqueries {
		if vs.AST == sub {
			validatedSub = vs
			break
		}
	}
	if validatedSub == nil {
		return "", fmt.Errorf("subquery not validated: %s", sub.From)
	}

	// Create nested shape
	nestedShape := &ResultShape{
		Object: validatedSub.ChildObject.Name,
		Table:  validatedSub.ChildObject.QualifiedTableName(),
	}

	// Build subquery SELECT
	childAlias := "sq"
	var selectParts []string

	for i, sel := range sub.Select {
		if sel.Item == nil || sel.Item.Expr == nil {
			continue
		}

		// For subqueries, we compile the expression in child context
		fieldExpr, err := c.compileSubqueryField(validatedSub.ChildObject, childAlias, sel.Item.Expr)
		if err != nil {
			return "", err
		}

		alias := sel.Alias
		if alias == nil {
			defaultAlias := fmt.Sprintf("f%d", i)
			alias = &defaultAlias
		}

		selectParts = append(selectParts, fmt.Sprintf("'%s', %s", *alias, fieldExpr))

		nestedShape.Fields = append(nestedShape.Fields, &FieldShape{
			Name:  *alias,
			Type:  sel.Item.Expr.FieldType,
			Alias: *alias,
		})
	}

	if len(selectParts) == 0 {
		return "NULL", nil
	}

	// Build the JSON_AGG subquery
	var sql strings.Builder
	sql.WriteString("(SELECT COALESCE(JSON_AGG(json_build_object(")
	sql.WriteString(strings.Join(selectParts, ", "))
	sql.WriteString(") ORDER BY ")

	// Add ORDER BY for subquery
	if len(sub.OrderBy) > 0 {
		orderParts, err := c.compileSubqueryOrderBy(validatedSub.ChildObject, childAlias, sub.OrderBy)
		if err != nil {
			return "", err
		}
		sql.WriteString(orderParts)
	} else {
		// Default ordering - use first field
		if len(selectParts) > 0 {
			sql.WriteString("1")
		}
	}

	sql.WriteString("), '[]'::json)")
	sql.WriteString(" FROM ")
	sql.WriteString(validatedSub.ChildObject.QualifiedTableName())
	sql.WriteString(" AS ")
	sql.WriteString(childAlias)
	sql.WriteString(" WHERE ")
	// ChildField and ParentField are SQL column names, not SOQL field names
	sql.WriteString(qualifiedColumn(childAlias, validatedSub.Relationship.ChildField))
	sql.WriteString(" = ")
	sql.WriteString(qualifiedColumn(ctx.mainAlias, validatedSub.Relationship.ParentField))

	// Add WHERE condition if present
	if sub.Where != nil {
		whereSQL, err := c.compileSubqueryExpression(ctx, validatedSub.ChildObject, childAlias, sub.Where)
		if err != nil {
			return "", err
		}
		sql.WriteString(" AND ")
		sql.WriteString(whereSQL)
	}

	// Add LIMIT if present
	limit := c.limits.EffectiveSubqueryLimit(sub.Limit)
	if limit > 0 {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}

	sql.WriteString(")")

	// Add relationship shape
	ctx.shape.Relationships = append(ctx.shape.Relationships, &RelationshipShape{
		Name:  sub.From,
		Shape: nestedShape,
	})

	return sql.String(), nil
}

// compileSubqueryField compiles a field expression within a subquery context.
func (c *Compiler) compileSubqueryField(obj *ObjectMeta, alias string, expr *Expression) (string, error) {
	if expr == nil || expr.Or == nil {
		return "", fmt.Errorf("empty expression")
	}
	return c.compileSubqueryOrExpr(obj, alias, expr.Or)
}

func (c *Compiler) compileSubqueryOrExpr(obj *ObjectMeta, alias string, or *OrExpr) (string, error) {
	if len(or.And) == 1 {
		return c.compileSubqueryAndExpr(obj, alias, or.And[0])
	}

	var parts []string
	for _, and := range or.And {
		part, err := c.compileSubqueryAndExpr(obj, alias, and)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " OR ") + ")", nil
}

func (c *Compiler) compileSubqueryAndExpr(obj *ObjectMeta, alias string, and *AndExpr) (string, error) {
	if len(and.Not) == 1 {
		return c.compileSubqueryNotExpr(obj, alias, and.Not[0])
	}

	var parts []string
	for _, not := range and.Not {
		part, err := c.compileSubqueryNotExpr(obj, alias, not)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func (c *Compiler) compileSubqueryNotExpr(obj *ObjectMeta, alias string, not *NotExpr) (string, error) {
	expr, err := c.compileSubqueryCompareExpr(obj, alias, not.Compare)
	if err != nil {
		return "", err
	}
	if not.Not {
		return "NOT " + expr, nil
	}
	return expr, nil
}

func (c *Compiler) compileSubqueryCompareExpr(obj *ObjectMeta, alias string, cmp *CompareExpr) (string, error) {
	left, err := c.compileSubqueryInExpr(obj, alias, cmp.Left)
	if err != nil {
		return "", err
	}

	if cmp.Operator == nil || cmp.Right == nil {
		return left, nil
	}

	right, err := c.compileSubqueryInExpr(obj, alias, cmp.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s", left, cmp.Operator.String(), right), nil
}

func (c *Compiler) compileSubqueryInExpr(obj *ObjectMeta, alias string, in *InExpr) (string, error) {
	left, err := c.compileSubqueryLikeExpr(obj, alias, in.Left)
	if err != nil {
		return "", err
	}
	// Simplified - not handling IN clause in subquery for now
	return left, nil
}

func (c *Compiler) compileSubqueryLikeExpr(obj *ObjectMeta, alias string, like *LikeExpr) (string, error) {
	left, err := c.compileSubqueryIsExpr(obj, alias, like.Left)
	if err != nil {
		return "", err
	}
	// Simplified - not handling LIKE in subquery for now
	return left, nil
}

func (c *Compiler) compileSubqueryIsExpr(obj *ObjectMeta, alias string, is *IsExpr) (string, error) {
	left, err := c.compileSubqueryAddExpr(obj, alias, is.Left)
	if err != nil {
		return "", err
	}

	if is.Is {
		if is.Not {
			return left + " IS NOT NULL", nil
		}
		return left + " IS NULL", nil
	}

	return left, nil
}

func (c *Compiler) compileSubqueryAddExpr(obj *ObjectMeta, alias string, add *AddExpr) (string, error) {
	return c.compileSubqueryMulExpr(obj, alias, add.Left)
}

func (c *Compiler) compileSubqueryMulExpr(obj *ObjectMeta, alias string, mul *MulExpr) (string, error) {
	return c.compileSubqueryUnaryExpr(obj, alias, mul.Left)
}

func (c *Compiler) compileSubqueryUnaryExpr(obj *ObjectMeta, alias string, unary *UnaryExpr) (string, error) {
	return c.compileSubqueryPrimary(obj, alias, unary.Primary)
}

func (c *Compiler) compileSubqueryPrimary(obj *ObjectMeta, alias string, primary *Primary) (string, error) {
	if primary == nil {
		return "", fmt.Errorf("empty primary")
	}

	switch {
	case primary.Field != nil:
		return c.compileSubqueryFieldRef(obj, alias, primary.Field)
	case primary.Const != nil:
		return c.compileConstValue(primary.Const), nil
	case primary.Subexpression != nil:
		inner, err := c.compileSubqueryField(obj, alias, primary.Subexpression)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil
	default:
		return "", fmt.Errorf("unsupported primary in subquery")
	}
}

func (c *Compiler) compileSubqueryFieldRef(obj *ObjectMeta, alias string, field *Field) (string, error) {
	if len(field.Path) == 0 {
		return "", fmt.Errorf("empty field path")
	}

	// Simple field reference (no lookups in subqueries for now)
	if len(field.Path) == 1 {
		fieldMeta := obj.GetField(field.Path[0])
		if fieldMeta == nil {
			return "", fmt.Errorf("unknown field: %s", field.Path[0])
		}
		return qualifiedColumn(alias, fieldMeta.Column), nil
	}

	// Dot notation - would need JOIN handling
	return "", fmt.Errorf("lookup paths in subqueries not yet supported")
}

// compileSubqueryOrderBy compiles ORDER BY for a subquery.
func (c *Compiler) compileSubqueryOrderBy(obj *ObjectMeta, alias string, orders []*OrderClause) (string, error) {
	var parts []string

	for _, order := range orders {
		if order.OrderItem == nil {
			continue
		}

		var fieldExpr string
		if order.OrderItem.Aggregate != nil {
			// Aggregates in subquery ORDER BY
			continue
		} else if len(order.OrderItem.Field) > 0 {
			fieldMeta := obj.GetField(order.OrderItem.Field[0])
			if fieldMeta == nil {
				continue
			}
			fieldExpr = qualifiedColumn(alias, fieldMeta.Column)
		} else {
			continue
		}

		part := fieldExpr
		if order.Direction != nil && *order.Direction == DirDesc {
			part += " DESC"
		}
		if order.Nulls != nil {
			part += " " + order.Nulls.String()
		}

		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return "1", nil
	}

	return strings.Join(parts, ", "), nil
}

// compileSubqueryExpression compiles an expression within a subquery context.
func (c *Compiler) compileSubqueryExpression(ctx *compileContext, obj *ObjectMeta, alias string, expr *Expression) (string, error) {
	return c.compileSubqueryField(obj, alias, expr)
}

// compileExpression compiles a general expression.
func (c *Compiler) compileExpression(ctx *compileContext, expr *Expression) (string, error) {
	if expr == nil || expr.Or == nil {
		return "", fmt.Errorf("empty expression")
	}
	return c.compileOrExpr(ctx, expr.Or)
}

func (c *Compiler) compileOrExpr(ctx *compileContext, or *OrExpr) (string, error) {
	if len(or.And) == 1 {
		return c.compileAndExpr(ctx, or.And[0])
	}

	var parts []string
	for _, and := range or.And {
		part, err := c.compileAndExpr(ctx, and)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " OR ") + ")", nil
}

func (c *Compiler) compileAndExpr(ctx *compileContext, and *AndExpr) (string, error) {
	if len(and.Not) == 1 {
		return c.compileNotExpr(ctx, and.Not[0])
	}

	var parts []string
	for _, not := range and.Not {
		part, err := c.compileNotExpr(ctx, not)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func (c *Compiler) compileNotExpr(ctx *compileContext, not *NotExpr) (string, error) {
	expr, err := c.compileCompareExpr(ctx, not.Compare)
	if err != nil {
		return "", err
	}
	if not.Not {
		return "NOT " + expr, nil
	}
	return expr, nil
}

func (c *Compiler) compileCompareExpr(ctx *compileContext, cmp *CompareExpr) (string, error) {
	left, err := c.compileInExpr(ctx, cmp.Left)
	if err != nil {
		return "", err
	}

	if cmp.Operator == nil || cmp.Right == nil {
		return left, nil
	}

	right, err := c.compileInExpr(ctx, cmp.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s", left, cmp.Operator.String(), right), nil
}

func (c *Compiler) compileInExpr(ctx *compileContext, in *InExpr) (string, error) {
	left, err := c.compileLikeExpr(ctx, in.Left)
	if err != nil {
		return "", err
	}

	if !in.In {
		return left, nil
	}

	op := "IN"
	if in.Not {
		op = "NOT IN"
	}

	// Handle WHERE subquery
	if in.Subquery != nil {
		subSQL, err := c.compileWhereSubquery(ctx, in.Subquery)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %s %s", left, op, subSQL), nil
	}

	// Compile IN values
	var values []string
	for _, val := range in.Values {
		v, err := c.compileValue(ctx, val)
		if err != nil {
			return "", err
		}
		values = append(values, v)
	}

	return fmt.Sprintf("%s %s (%s)", left, op, strings.Join(values, ", ")), nil
}

// compileWhereSubquery compiles a WHERE IN subquery to SQL.
func (c *Compiler) compileWhereSubquery(ctx *compileContext, sub *WhereSubquery) (string, error) {
	// Find validated subquery
	var validated *ValidatedWhereSubquery
	for _, vs := range ctx.whereSubqueries {
		if vs.AST == sub {
			validated = vs
			break
		}
	}
	if validated == nil {
		return "", fmt.Errorf("WHERE subquery not validated: %s", sub.From)
	}

	alias := "wsq"

	// Build: (SELECT field FROM table AS wsq WHERE ...)
	var sql strings.Builder
	sql.WriteString("(SELECT ")
	sql.WriteString(qualifiedColumn(alias, validated.Field.Column))
	sql.WriteString(" FROM ")
	sql.WriteString(validated.Object.QualifiedTableName())
	sql.WriteString(" AS ")
	sql.WriteString(alias)

	if sub.Where != nil {
		whereSQL, err := c.compileSubqueryField(validated.Object, alias, sub.Where)
		if err != nil {
			return "", err
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereSQL)
	}

	if sub.Limit != nil {
		sql.WriteString(fmt.Sprintf(" LIMIT %d", *sub.Limit))
	}

	sql.WriteString(")")
	return sql.String(), nil
}

func (c *Compiler) compileLikeExpr(ctx *compileContext, like *LikeExpr) (string, error) {
	left, err := c.compileIsExpr(ctx, like.Left)
	if err != nil {
		return "", err
	}

	if !like.Like || like.Pattern == nil {
		return left, nil
	}

	pattern, err := c.compileValue(ctx, like.Pattern)
	if err != nil {
		return "", err
	}

	op := "LIKE"
	if like.Not {
		op = "NOT LIKE"
	}

	return fmt.Sprintf("%s %s %s", left, op, pattern), nil
}

func (c *Compiler) compileIsExpr(ctx *compileContext, is *IsExpr) (string, error) {
	left, err := c.compileAddExpr(ctx, is.Left)
	if err != nil {
		return "", err
	}

	if !is.Is {
		return left, nil
	}

	if is.Not {
		return left + " IS NOT NULL", nil
	}
	return left + " IS NULL", nil
}

func (c *Compiler) compileAddExpr(ctx *compileContext, add *AddExpr) (string, error) {
	result, err := c.compileMulExpr(ctx, add.Left)
	if err != nil {
		return "", err
	}

	for _, op := range add.Right {
		right, err := c.compileMulExpr(ctx, op.Right)
		if err != nil {
			return "", err
		}
		result = fmt.Sprintf("%s %s %s", result, op.Operator.String(), right)
	}

	return result, nil
}

func (c *Compiler) compileMulExpr(ctx *compileContext, mul *MulExpr) (string, error) {
	result, err := c.compileUnaryExpr(ctx, mul.Left)
	if err != nil {
		return "", err
	}

	for _, op := range mul.Right {
		right, err := c.compileUnaryExpr(ctx, op.Right)
		if err != nil {
			return "", err
		}
		result = fmt.Sprintf("%s %s %s", result, op.Operator.String(), right)
	}

	return result, nil
}

func (c *Compiler) compileUnaryExpr(ctx *compileContext, unary *UnaryExpr) (string, error) {
	primary, err := c.compilePrimary(ctx, unary.Primary)
	if err != nil {
		return "", err
	}

	if unary.Operator != nil {
		return fmt.Sprintf("%s%s", unary.Operator.String(), primary), nil
	}
	return primary, nil
}

func (c *Compiler) compilePrimary(ctx *compileContext, primary *Primary) (string, error) {
	if primary == nil {
		return "", fmt.Errorf("empty primary")
	}

	switch {
	case primary.Subexpression != nil:
		inner, err := c.compileExpression(ctx, primary.Subexpression)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil

	case primary.Aggregate != nil:
		return c.compileAggregate(ctx, primary.Aggregate)

	case primary.FuncCall != nil:
		return c.compileFuncCall(ctx, primary.FuncCall)

	case primary.Const != nil:
		return c.compileConst(ctx, primary.Const), nil

	case primary.Field != nil:
		return c.compileField(ctx, primary.Field)

	default:
		return "", fmt.Errorf("invalid primary expression")
	}
}

func (c *Compiler) compileValue(ctx *compileContext, val *Value) (string, error) {
	if val == nil {
		return "", fmt.Errorf("empty value")
	}

	if val.Const != nil {
		return c.compileConst(ctx, val.Const), nil
	}

	if val.Field != nil {
		return c.compileField(ctx, val.Field)
	}

	return "", fmt.Errorf("invalid value")
}

func (c *Compiler) compileFuncCall(ctx *compileContext, fn *FuncCall) (string, error) {
	var args []string
	for _, arg := range fn.Args {
		a, err := c.compileExpression(ctx, arg)
		if err != nil {
			return "", err
		}
		args = append(args, a)
	}

	// Map SOQL function names to PostgreSQL function names
	switch fn.Name {
	case FuncCoalesce:
		return fmt.Sprintf("COALESCE(%s)", strings.Join(args, ", ")), nil
	case FuncNullif:
		return fmt.Sprintf("NULLIF(%s)", strings.Join(args, ", ")), nil
	case FuncConcat:
		return fmt.Sprintf("CONCAT(%s)", strings.Join(args, ", ")), nil
	case FuncUpper:
		return fmt.Sprintf("UPPER(%s)", args[0]), nil
	case FuncLower:
		return fmt.Sprintf("LOWER(%s)", args[0]), nil
	case FuncTrim:
		return fmt.Sprintf("TRIM(%s)", args[0]), nil
	case FuncLength:
		return fmt.Sprintf("LENGTH(%s)", args[0]), nil
	case FuncSubstring:
		// PostgreSQL SUBSTRING(string, start [, length])
		return fmt.Sprintf("SUBSTRING(%s)", strings.Join(args, ", ")), nil
	case FuncAbs:
		return fmt.Sprintf("ABS(%s)", args[0]), nil
	case FuncRound:
		return fmt.Sprintf("ROUND(%s)", strings.Join(args, ", ")), nil
	case FuncFloor:
		return fmt.Sprintf("FLOOR(%s)", args[0]), nil
	case FuncCeil:
		return fmt.Sprintf("CEIL(%s)", args[0]), nil
	default:
		return "", fmt.Errorf("unsupported function: %s", fn.Name)
	}
}

// compileConst compiles a constant value, adding parameters as needed.
func (c *Compiler) compileConst(ctx *compileContext, cnst *Const) string {
	// Handle date literals specially - they need runtime resolution
	if cnst.StaticDate != nil || cnst.DynamicDate != nil {
		ctx.paramCount++
		paramIdx := ctx.paramCount
		ctx.params = append(ctx.params, nil) // Placeholder

		dp := &DateParam{
			ParamIndex: paramIdx,
		}
		if cnst.StaticDate != nil {
			dp.Static = cnst.StaticDate
		}
		if cnst.DynamicDate != nil {
			dp.Dynamic = cnst.DynamicDate
		}
		ctx.dateParams = append(ctx.dateParams, dp)

		return fmt.Sprintf("$%d", paramIdx)
	}

	return c.compileConstValue(cnst)
}

// compileConstValue returns the SQL representation of a constant.
func (c *Compiler) compileConstValue(cnst *Const) string {
	switch {
	case cnst.Null:
		return "NULL"
	case cnst.String != nil:
		// Escape single quotes
		escaped := strings.ReplaceAll(*cnst.String, "'", "''")
		return "'" + escaped + "'"
	case cnst.Integer != nil:
		return fmt.Sprintf("%d", *cnst.Integer)
	case cnst.Float != nil:
		return fmt.Sprintf("%g", *cnst.Float)
	case cnst.Boolean != nil:
		if bool(*cnst.Boolean) {
			return "TRUE"
		}
		return "FALSE"
	case cnst.Date != nil:
		return "'" + cnst.Date.Format("2006-01-02") + "'"
	case cnst.DateTime != nil:
		return "'" + cnst.DateTime.Format("2006-01-02T15:04:05Z07:00") + "'"
	default:
		return "NULL"
	}
}

// compileField compiles a field reference.
func (c *Compiler) compileField(ctx *compileContext, field *Field) (string, error) {
	if field == nil || len(field.Path) == 0 {
		return "", fmt.Errorf("empty field reference")
	}

	pathKey := strings.Join(field.Path, ".")

	// Check if this is a resolved reference
	if ref, ok := ctx.validated.ResolvedRefs[pathKey]; ok {
		// If there are joins, use the join alias
		if len(ref.Joins) > 0 {
			alias := c.ensureJoin(ctx, ref.Joins)
			return qualifiedColumn(alias, ref.Field.Column), nil
		}

		// Direct field on root object
		return qualifiedColumn(ctx.mainAlias, ref.Field.Column), nil
	}

	// Fallback - try to resolve from root object
	if len(field.Path) == 1 {
		fieldMeta := ctx.validated.RootObject.GetField(field.Path[0])
		if fieldMeta != nil {
			return qualifiedColumn(ctx.mainAlias, fieldMeta.Column), nil
		}
	}

	return "", fmt.Errorf("unresolved field: %s", pathKey)
}

// buildJoins builds JOIN clauses from resolved references.
func (c *Compiler) buildJoins(ctx *compileContext) error {
	// Collect unique join paths
	for _, ref := range ctx.validated.ResolvedRefs {
		if len(ref.Joins) > 0 {
			c.ensureJoin(ctx, ref.Joins)
		}
	}
	return nil
}

// ensureJoin ensures that the required joins exist and returns the final alias.
func (c *Compiler) ensureJoin(ctx *compileContext, joins []*Join) string {
	var prevAlias = ctx.mainAlias

	for _, join := range joins {
		// Build unique key for this join
		joinKey := join.Alias

		if existingAlias, ok := ctx.joinAliases[joinKey]; ok {
			prevAlias = existingAlias
			continue
		}

		// Create new alias
		newAlias := fmt.Sprintf("t%d", len(ctx.joinAliases)+1)
		ctx.joinAliases[joinKey] = newAlias

		// Get column names
		fromColumn := join.FromObject.GetField(join.FromField)
		if fromColumn == nil {
			// Try to find it as the direct field name
			fromColumn = &FieldMeta{Column: join.FromField}
		}

		toColumn := join.ToObject.GetField(join.ToField)
		if toColumn == nil {
			toColumn = &FieldMeta{Column: join.ToField}
		}

		// Build JOIN SQL
		joinSQL := fmt.Sprintf("LEFT JOIN %s AS %s ON %s = %s",
			join.ToObject.QualifiedTableName(),
			newAlias,
			qualifiedColumn(prevAlias, fromColumn.Column),
			qualifiedColumn(newAlias, toColumn.Column),
		)

		ctx.joinSQL = append(ctx.joinSQL, joinSQL)
		prevAlias = newAlias
	}

	return prevAlias
}

// compileGroupBy compiles the GROUP BY clause.
func (c *Compiler) compileGroupBy(ctx *compileContext, groups []*GroupClause) (string, error) {
	var parts []string

	for _, group := range groups {
		pathKey := strings.Join(group.Field, ".")

		if ref, ok := ctx.validated.ResolvedRefs[pathKey]; ok {
			if len(ref.Joins) > 0 {
				alias := c.ensureJoin(ctx, ref.Joins)
				parts = append(parts, qualifiedColumn(alias, ref.Field.Column))
			} else {
				parts = append(parts, qualifiedColumn(ctx.mainAlias, ref.Field.Column))
			}
		} else if len(group.Field) == 1 {
			fieldMeta := ctx.validated.RootObject.GetField(group.Field[0])
			if fieldMeta != nil {
				parts = append(parts, qualifiedColumn(ctx.mainAlias, fieldMeta.Column))
			}
		}
	}

	return strings.Join(parts, ", "), nil
}

// compileOrderBy compiles the ORDER BY clause and tracks keyset fields.
func (c *Compiler) compileOrderBy(ctx *compileContext, orders []*OrderClause) (string, error) {
	var parts []string

	for _, order := range orders {
		if order.OrderItem == nil {
			continue
		}

		var fieldExpr string
		var tableAlias string
		var columnName string
		var soqlName string
		var err error

		if order.OrderItem.Aggregate != nil {
			// Aggregates can't be used for keyset pagination
			fieldExpr, err = c.compileAggregate(ctx, order.OrderItem.Aggregate)
			if err != nil {
				return "", err
			}
		} else if len(order.OrderItem.Field) > 0 {
			pathKey := strings.Join(order.OrderItem.Field, ".")
			soqlName = pathKey

			if ref, ok := ctx.validated.ResolvedRefs[pathKey]; ok {
				if len(ref.Joins) > 0 {
					tableAlias = c.ensureJoin(ctx, ref.Joins)
					columnName = ref.Field.Column
					fieldExpr = qualifiedColumn(tableAlias, columnName)
				} else {
					tableAlias = ctx.mainAlias
					columnName = ref.Field.Column
					fieldExpr = qualifiedColumn(tableAlias, columnName)
				}
			} else if len(order.OrderItem.Field) == 1 {
				fieldMeta := ctx.validated.RootObject.GetField(order.OrderItem.Field[0])
				if fieldMeta != nil {
					tableAlias = ctx.mainAlias
					columnName = fieldMeta.Column
					fieldExpr = qualifiedColumn(tableAlias, columnName)
				}
			}
		}

		if fieldExpr == "" {
			continue
		}

		// Determine direction
		direction := "asc"
		part := fieldExpr
		if order.Direction != nil && *order.Direction == DirDesc {
			direction = "desc"
			part += " DESC"
		}
		if order.Nulls != nil && *order.Nulls != NullsDefault {
			part += " " + order.Nulls.String()
		}

		parts = append(parts, part)

		// Track keyset field (only for regular fields, not aggregates)
		if columnName != "" {
			ctx.keysetFields = append(ctx.keysetFields, &KeysetField{
				SOQLName:   soqlName,
				SQLColumn:  columnName,
				TableAlias: tableAlias,
				Direction:  direction,
			})
		}
	}

	return strings.Join(parts, ", "), nil
}
