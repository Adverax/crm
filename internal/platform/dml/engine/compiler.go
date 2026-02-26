package engine

import (
	"fmt"
	"strings"
)

// CompiledDML represents a compiled DML statement ready for execution.
type CompiledDML struct {
	// SQL is the generated SQL with placeholders ($1, $2...)
	SQL string

	// Params are the parameter values in order
	Params []any

	// Operation is the type of DML operation
	Operation Operation

	// Object is the target object name
	Object string

	// Table is the target SQL table name
	Table string

	// RowCount is the number of rows affected (for INSERT/UPSERT)
	RowCount int

	// ReturningColumn is the column name to return (usually id)
	ReturningColumn string
}

// Compiler compiles validated DML statements to SQL.
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
	compiler   *Compiler
	validated  *ValidatedDML
	params     []any
	paramCount int
}

func newCompileContext(c *Compiler, v *ValidatedDML) *compileContext {
	return &compileContext{
		compiler:   c,
		validated:  v,
		params:     make([]any, 0),
		paramCount: 0,
	}
}

// addParam adds a parameter and returns its placeholder ($1, $2, etc.)
func (ctx *compileContext) addParam(value any) string {
	ctx.paramCount++
	ctx.params = append(ctx.params, value)
	return fmt.Sprintf("$%d", ctx.paramCount)
}

// Compile compiles a validated DML statement to SQL.
func (c *Compiler) Compile(validated *ValidatedDML) (*CompiledDML, error) {
	switch validated.Operation {
	case OperationInsert:
		return c.compileInsert(validated)
	case OperationUpdate:
		return c.compileUpdate(validated)
	case OperationDelete:
		return c.compileDelete(validated)
	case OperationUpsert:
		return c.compileUpsert(validated)
	default:
		return nil, NewValidationError(ErrCodeInvalidExpression, "unknown operation")
	}
}

// compileInsert compiles an INSERT statement.
// Generates: INSERT INTO table (col1, col2) VALUES ($1, $2), ($3, $4) RETURNING id
func (c *Compiler) compileInsert(validated *ValidatedDML) (*CompiledDML, error) {
	ctx := newCompileContext(c, validated)
	ins := validated.AST.Insert

	var sql strings.Builder

	// INSERT INTO table
	sql.WriteString("INSERT INTO ")
	sql.WriteString(validated.Object.Table())

	// (col1, col2, ...)
	sql.WriteString(" (")
	columns := make([]string, len(validated.Fields))
	for i, f := range validated.Fields {
		columns[i] = f.Column
	}
	sql.WriteString(strings.Join(columns, ", "))
	sql.WriteString(")")

	// VALUES ($1, $2), ($3, $4), ...
	sql.WriteString(" VALUES ")
	rowPlaceholders := make([]string, len(ins.Values))
	for i, row := range ins.Values {
		valuePlaceholders := make([]string, len(row.Values))
		for j, val := range row.Values {
			compiled, err := c.compileExpr(ctx, validated.Object, val)
			if err != nil {
				return nil, err
			}
			valuePlaceholders[j] = compiled
		}
		rowPlaceholders[i] = "(" + strings.Join(valuePlaceholders, ", ") + ")"
	}
	sql.WriteString(strings.Join(rowPlaceholders, ", "))

	// RETURNING id
	sql.WriteString(" RETURNING ")
	sql.WriteString(validated.Object.PrimaryKey)

	return &CompiledDML{
		SQL:             sql.String(),
		Params:          ctx.params,
		Operation:       OperationInsert,
		Object:          validated.Object.Name,
		Table:           validated.Object.Table(),
		RowCount:        len(ins.Values),
		ReturningColumn: validated.Object.PrimaryKey,
	}, nil
}

// compileUpdate compiles an UPDATE statement.
// Generates: UPDATE table SET col1 = $1, col2 = $2 WHERE ... RETURNING id
func (c *Compiler) compileUpdate(validated *ValidatedDML) (*CompiledDML, error) {
	ctx := newCompileContext(c, validated)
	upd := validated.AST.Update

	var sql strings.Builder

	// UPDATE table
	sql.WriteString("UPDATE ")
	sql.WriteString(validated.Object.Table())

	// SET col1 = $1, col2 = $2
	sql.WriteString(" SET ")
	setParts := make([]string, len(validated.Assignments))
	for i, assign := range validated.Assignments {
		compiled, err := c.compileExpr(ctx, validated.Object, assign.Value)
		if err != nil {
			return nil, err
		}
		setParts[i] = fmt.Sprintf("%s = %s", assign.Field.Column, compiled)
	}
	sql.WriteString(strings.Join(setParts, ", "))

	// WHERE ...
	if upd.Where != nil {
		whereSQL, err := c.compileExpression(ctx, validated.Object, upd.Where)
		if err != nil {
			return nil, err
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereSQL)
	}

	// RETURNING id
	sql.WriteString(" RETURNING ")
	sql.WriteString(validated.Object.PrimaryKey)

	return &CompiledDML{
		SQL:             sql.String(),
		Params:          ctx.params,
		Operation:       OperationUpdate,
		Object:          validated.Object.Name,
		Table:           validated.Object.Table(),
		ReturningColumn: validated.Object.PrimaryKey,
	}, nil
}

// compileDelete compiles a DELETE statement.
// Generates: DELETE FROM table WHERE ... RETURNING id
func (c *Compiler) compileDelete(validated *ValidatedDML) (*CompiledDML, error) {
	ctx := newCompileContext(c, validated)
	del := validated.AST.Delete

	var sql strings.Builder

	// DELETE FROM table
	sql.WriteString("DELETE FROM ")
	sql.WriteString(validated.Object.Table())

	// WHERE ...
	if del.Where != nil {
		whereSQL, err := c.compileExpression(ctx, validated.Object, del.Where)
		if err != nil {
			return nil, err
		}
		sql.WriteString(" WHERE ")
		sql.WriteString(whereSQL)
	}

	// RETURNING id
	sql.WriteString(" RETURNING ")
	sql.WriteString(validated.Object.PrimaryKey)

	return &CompiledDML{
		SQL:             sql.String(),
		Params:          ctx.params,
		Operation:       OperationDelete,
		Object:          validated.Object.Name,
		Table:           validated.Object.Table(),
		ReturningColumn: validated.Object.PrimaryKey,
	}, nil
}

// compileUpsert compiles an UPSERT statement.
// Generates: INSERT INTO table (col1, col2) VALUES ($1, $2)
//
//	ON CONFLICT (external_id) DO UPDATE SET col1 = EXCLUDED.col1, ...
//	RETURNING id
func (c *Compiler) compileUpsert(validated *ValidatedDML) (*CompiledDML, error) {
	ctx := newCompileContext(c, validated)
	ups := validated.AST.Upsert

	var sql strings.Builder

	// INSERT INTO table
	sql.WriteString("INSERT INTO ")
	sql.WriteString(validated.Object.Table())

	// (col1, col2, ...)
	sql.WriteString(" (")
	columns := make([]string, len(validated.Fields))
	for i, f := range validated.Fields {
		columns[i] = f.Column
	}
	sql.WriteString(strings.Join(columns, ", "))
	sql.WriteString(")")

	// VALUES ($1, $2), ($3, $4), ...
	sql.WriteString(" VALUES ")
	rowPlaceholders := make([]string, len(ups.Values))
	for i, row := range ups.Values {
		valuePlaceholders := make([]string, len(row.Values))
		for j, val := range row.Values {
			compiled, err := c.compileExpr(ctx, validated.Object, val)
			if err != nil {
				return nil, err
			}
			valuePlaceholders[j] = compiled
		}
		rowPlaceholders[i] = "(" + strings.Join(valuePlaceholders, ", ") + ")"
	}
	sql.WriteString(strings.Join(rowPlaceholders, ", "))

	// ON CONFLICT (external_id)
	sql.WriteString(" ON CONFLICT (")
	sql.WriteString(validated.ExternalIdField.Column)
	sql.WriteString(")")

	// DO UPDATE SET col1 = EXCLUDED.col1, col2 = EXCLUDED.col2, ...
	// Exclude the external ID field from updates
	sql.WriteString(" DO UPDATE SET ")
	updateParts := make([]string, 0, len(validated.Fields)-1)
	for _, f := range validated.Fields {
		if f.Name == validated.ExternalIdField.Name {
			continue // Don't update the conflict key
		}
		updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", f.Column, f.Column))
	}
	if len(updateParts) > 0 {
		sql.WriteString(strings.Join(updateParts, ", "))
	} else {
		// If only the external ID is provided, use a no-op update
		fmt.Fprintf(&sql, "%s = EXCLUDED.%s",
			validated.ExternalIdField.Column, validated.ExternalIdField.Column)
	}

	// RETURNING id
	sql.WriteString(" RETURNING ")
	sql.WriteString(validated.Object.PrimaryKey)

	return &CompiledDML{
		SQL:             sql.String(),
		Params:          ctx.params,
		Operation:       OperationUpsert,
		Object:          validated.Object.Name,
		Table:           validated.Object.Table(),
		RowCount:        len(ups.Values),
		ReturningColumn: validated.Object.PrimaryKey,
	}, nil
}

// compileExpression compiles a WHERE expression to SQL.
func (c *Compiler) compileExpression(ctx *compileContext, obj *ObjectMeta, expr *Expression) (string, error) {
	if expr == nil || expr.Or == nil {
		return "", fmt.Errorf("empty expression")
	}
	return c.compileOrExpr(ctx, obj, expr.Or)
}

func (c *Compiler) compileOrExpr(ctx *compileContext, obj *ObjectMeta, or *OrExpr) (string, error) {
	if len(or.And) == 1 {
		return c.compileAndExpr(ctx, obj, or.And[0])
	}

	var parts []string
	for _, and := range or.And {
		part, err := c.compileAndExpr(ctx, obj, and)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " OR ") + ")", nil
}

func (c *Compiler) compileAndExpr(ctx *compileContext, obj *ObjectMeta, and *AndExpr) (string, error) {
	if len(and.Not) == 1 {
		return c.compileNotExpr(ctx, obj, and.Not[0])
	}

	var parts []string
	for _, not := range and.Not {
		part, err := c.compileNotExpr(ctx, obj, not)
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}
	return "(" + strings.Join(parts, " AND ") + ")", nil
}

func (c *Compiler) compileNotExpr(ctx *compileContext, obj *ObjectMeta, not *NotExpr) (string, error) {
	expr, err := c.compileCompareExpr(ctx, obj, not.Compare)
	if err != nil {
		return "", err
	}
	if not.Not {
		return "NOT " + expr, nil
	}
	return expr, nil
}

func (c *Compiler) compileCompareExpr(ctx *compileContext, obj *ObjectMeta, cmp *CompareExpr) (string, error) {
	left, err := c.compileInExpr(ctx, obj, cmp.Left)
	if err != nil {
		return "", err
	}

	if cmp.Operator == nil || cmp.Right == nil {
		return left, nil
	}

	right, err := c.compileInExpr(ctx, obj, cmp.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s %s", left, cmp.Operator.String(), right), nil
}

func (c *Compiler) compileInExpr(ctx *compileContext, obj *ObjectMeta, in *InExpr) (string, error) {
	left, err := c.compileLikeExpr(ctx, obj, in.Left)
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

	// Compile IN values
	var values []string
	for _, val := range in.Values {
		v, err := c.compileValue(ctx, obj, val)
		if err != nil {
			return "", err
		}
		values = append(values, v)
	}

	return fmt.Sprintf("%s %s (%s)", left, op, strings.Join(values, ", ")), nil
}

func (c *Compiler) compileLikeExpr(ctx *compileContext, obj *ObjectMeta, like *LikeExpr) (string, error) {
	left, err := c.compileIsExpr(ctx, obj, like.Left)
	if err != nil {
		return "", err
	}

	if !like.Like || like.Pattern == nil {
		return left, nil
	}

	pattern, err := c.compileValue(ctx, obj, like.Pattern)
	if err != nil {
		return "", err
	}

	op := "LIKE"
	if like.Not {
		op = "NOT LIKE"
	}

	return fmt.Sprintf("%s %s %s", left, op, pattern), nil
}

func (c *Compiler) compileIsExpr(ctx *compileContext, obj *ObjectMeta, is *IsExpr) (string, error) {
	left, err := c.compilePrimary(ctx, obj, is.Left)
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

func (c *Compiler) compilePrimary(ctx *compileContext, obj *ObjectMeta, primary *Primary) (string, error) {
	if primary == nil {
		return "", fmt.Errorf("empty primary")
	}

	switch {
	case primary.Subexpression != nil:
		inner, err := c.compileExpression(ctx, obj, primary.Subexpression)
		if err != nil {
			return "", err
		}
		return "(" + inner + ")", nil

	case primary.Const != nil:
		return c.compileConst(ctx, primary.Const), nil

	case primary.Field != nil:
		return c.compileField(ctx, obj, primary.Field)

	default:
		return "", fmt.Errorf("invalid primary expression")
	}
}

func (c *Compiler) compileValue(ctx *compileContext, obj *ObjectMeta, val *Value) (string, error) {
	if val == nil {
		return "", fmt.Errorf("empty value")
	}

	if val.Const != nil {
		return c.compileConst(ctx, val.Const), nil
	}

	if val.Field != nil {
		return c.compileField(ctx, obj, val.Field)
	}

	return "", fmt.Errorf("invalid value")
}

// compileExpr compiles a value expression (constant, function call, or field reference).
func (c *Compiler) compileExpr(ctx *compileContext, obj *ObjectMeta, expr *Expr) (string, error) {
	if expr == nil {
		return "", fmt.Errorf("empty expression")
	}

	switch {
	case expr.FuncCall != nil:
		return c.compileFuncCall(ctx, obj, expr.FuncCall)
	case expr.Const != nil:
		return c.compileConst(ctx, expr.Const), nil
	case expr.Field != nil:
		return c.compileField(ctx, obj, expr.Field)
	default:
		return "", fmt.Errorf("invalid expression")
	}
}

// compileFuncCall compiles a function call to SQL.
func (c *Compiler) compileFuncCall(ctx *compileContext, obj *ObjectMeta, fc *FuncCall) (string, error) {
	// Compile arguments
	args := make([]string, len(fc.Args))
	for i, arg := range fc.Args {
		compiled, err := c.compileExpr(ctx, obj, arg)
		if err != nil {
			return "", err
		}
		args[i] = compiled
	}

	// Build function call SQL
	return fmt.Sprintf("%s(%s)", fc.Name.String(), strings.Join(args, ", ")), nil
}

// compileConst compiles a constant value to a parameter placeholder.
func (c *Compiler) compileConst(ctx *compileContext, cnst *Const) string {
	if cnst.Null {
		return "NULL"
	}
	return ctx.addParam(cnst.Value())
}

// compileField compiles a field reference to SQL column name.
func (c *Compiler) compileField(ctx *compileContext, obj *ObjectMeta, field *Field) (string, error) {
	if field == nil || field.Name == "" {
		return "", fmt.Errorf("empty field reference")
	}

	fieldMeta := obj.GetField(field.Name)
	if fieldMeta == nil {
		return "", UnknownFieldError(obj.Name, field.Name)
	}

	return fieldMeta.Column, nil
}
