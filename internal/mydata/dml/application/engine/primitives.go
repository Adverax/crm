package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FieldType represents the type of a field value
type FieldType int

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeNull
	FieldTypeString
	FieldTypeInteger
	FieldTypeFloat
	FieldTypeBoolean
	FieldTypeDate
	FieldTypeDateTime
	FieldTypeID
)

func (t FieldType) String() string {
	switch t {
	case FieldTypeUnknown:
		return "unknown"
	case FieldTypeNull:
		return "null"
	case FieldTypeString:
		return "string"
	case FieldTypeInteger:
		return "integer"
	case FieldTypeFloat:
		return "float"
	case FieldTypeBoolean:
		return "boolean"
	case FieldTypeDate:
		return "date"
	case FieldTypeDateTime:
		return "datetime"
	case FieldTypeID:
		return "id"
	default:
		return "unknown"
	}
}

// IsCompatibleWith checks if this type can be assigned to the target type.
func (t FieldType) IsCompatibleWith(target FieldType) bool {
	if t == target {
		return true
	}
	// NULL is compatible with any nullable type
	if t == FieldTypeNull {
		return true
	}
	// String can be used for ID fields
	if t == FieldTypeString && target == FieldTypeID {
		return true
	}
	// Integer can be coerced to Float
	if t == FieldTypeInteger && target == FieldTypeFloat {
		return true
	}
	return false
}

// Operator represents comparison operators used in WHERE clauses
type Operator int

const (
	OpEQ Operator = iota // =
	OpNE                 // != or <>
	OpGT                 // >
	OpLT                 // <
	OpGE                 // >=
	OpLE                 // <=
)

func (op *Operator) Capture(s []string) error {
	val := strings.TrimSpace(s[0])
	switch val {
	case "=", "==":
		*op = OpEQ
	case "!=", "<>":
		*op = OpNE
	case ">":
		*op = OpGT
	case "<":
		*op = OpLT
	case ">=":
		*op = OpGE
	case "<=":
		*op = OpLE
	default:
		*op = OpEQ
	}
	return nil
}

func (op Operator) String() string {
	switch op {
	case OpEQ:
		return "="
	case OpNE:
		return "!="
	case OpGT:
		return ">"
	case OpLT:
		return "<"
	case OpGE:
		return ">="
	case OpLE:
		return "<="
	default:
		return "?"
	}
}

// Boolean is a custom boolean type for parsing
type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = strings.ToUpper(values[0]) == "TRUE"
	return nil
}

// Date represents a date value (YYYY-MM-DD)
type Date struct {
	time.Time
}

func (d *Date) Capture(s []string) error {
	t, err := time.Parse("2006-01-02", s[0])
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d Date) Format(layout string) string {
	return d.Time.Format(layout)
}

// DateTime represents a datetime value (YYYY-MM-DDTHH:MM:SSZ)
type DateTime struct {
	time.Time
}

func (d *DateTime) Capture(s []string) error {
	t, err := time.Parse(time.RFC3339, s[0])
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d DateTime) Format(layout string) string {
	return d.Time.Format(layout)
}

// Const represents a constant value in DML statements
type Const struct {
	DateTime  *DateTime `  @DateTime`
	Date      *Date     `| @Date`
	String    *string   `| @String`
	Float     *float64  `| @Float`
	Integer   *int      `| @Integer`
	Boolean   *Boolean  `| @("TRUE" | "FALSE")`
	Null      bool      `| @"NULL"`
	FieldType FieldType // Inferred type
}

// GetFieldType returns the inferred type of the constant
func (c *Const) GetFieldType() FieldType {
	if c.FieldType != FieldTypeUnknown {
		return c.FieldType
	}

	switch {
	case c.String != nil:
		c.FieldType = FieldTypeString
	case c.Integer != nil:
		c.FieldType = FieldTypeInteger
	case c.Float != nil:
		c.FieldType = FieldTypeFloat
	case c.Boolean != nil:
		c.FieldType = FieldTypeBoolean
	case c.Date != nil:
		c.FieldType = FieldTypeDate
	case c.DateTime != nil:
		c.FieldType = FieldTypeDateTime
	case c.Null:
		c.FieldType = FieldTypeNull
	default:
		c.FieldType = FieldTypeUnknown
	}

	return c.FieldType
}

// Value returns the Go value of the constant
func (c *Const) Value() any {
	switch {
	case c.String != nil:
		return *c.String
	case c.Integer != nil:
		return *c.Integer
	case c.Float != nil:
		return *c.Float
	case c.Boolean != nil:
		return bool(*c.Boolean)
	case c.Date != nil:
		return c.Date.Time
	case c.DateTime != nil:
		return c.DateTime.Time
	case c.Null:
		return nil
	default:
		return nil
	}
}

// SQLValue returns the SQL representation of the constant as a string
func (c *Const) SQLValue() string {
	switch {
	case c.Null:
		return "NULL"
	case c.String != nil:
		// Escape single quotes
		escaped := strings.ReplaceAll(*c.String, "'", "''")
		return "'" + escaped + "'"
	case c.Integer != nil:
		return strconv.Itoa(*c.Integer)
	case c.Float != nil:
		return fmt.Sprintf("%g", *c.Float)
	case c.Boolean != nil:
		if bool(*c.Boolean) {
			return "TRUE"
		}
		return "FALSE"
	case c.Date != nil:
		return "'" + c.Date.Format("2006-01-02") + "'"
	case c.DateTime != nil:
		return "'" + c.DateTime.Format("2006-01-02T15:04:05Z07:00") + "'"
	default:
		return "NULL"
	}
}

// NewStringConst creates a new string constant.
func NewStringConst(s string) *Const {
	return &Const{String: &s, FieldType: FieldTypeString}
}

// NewIntConst creates a new integer constant.
func NewIntConst(i int) *Const {
	return &Const{Integer: &i, FieldType: FieldTypeInteger}
}

// NewFloatConst creates a new float constant.
func NewFloatConst(f float64) *Const {
	return &Const{Float: &f, FieldType: FieldTypeFloat}
}

// NewBoolConst creates a new boolean constant.
func NewBoolConst(b bool) *Const {
	bb := Boolean(b)
	return &Const{Boolean: &bb, FieldType: FieldTypeBoolean}
}

// NewNullConst creates a new NULL constant.
func NewNullConst() *Const {
	return &Const{Null: true, FieldType: FieldTypeNull}
}

// NewDateConst creates a new date constant.
func NewDateConst(t time.Time) *Const {
	return &Const{Date: &Date{Time: t}, FieldType: FieldTypeDate}
}

// NewDateTimeConst creates a new datetime constant.
func NewDateTimeConst(t time.Time) *Const {
	return &Const{DateTime: &DateTime{Time: t}, FieldType: FieldTypeDateTime}
}

// =============================================================================
// Functions
// =============================================================================

// Function represents built-in scalar functions
type Function int

const (
	FuncCoalesce Function = iota
	FuncNullif
	FuncConcat
	FuncUpper
	FuncLower
	FuncTrim
	FuncLength
	FuncSubstring
	FuncAbs
	FuncRound
	FuncFloor
	FuncCeil
)

func (f *Function) Capture(s []string) error {
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	switch val {
	case "COALESCE":
		*f = FuncCoalesce
	case "NULLIF":
		*f = FuncNullif
	case "CONCAT":
		*f = FuncConcat
	case "UPPER":
		*f = FuncUpper
	case "LOWER":
		*f = FuncLower
	case "TRIM":
		*f = FuncTrim
	case "LENGTH", "LEN":
		*f = FuncLength
	case "SUBSTRING", "SUBSTR":
		*f = FuncSubstring
	case "ABS":
		*f = FuncAbs
	case "ROUND":
		*f = FuncRound
	case "FLOOR":
		*f = FuncFloor
	case "CEIL", "CEILING":
		*f = FuncCeil
	default:
		return fmt.Errorf("unknown function: %s", val)
	}
	return nil
}

func (f Function) String() string {
	switch f {
	case FuncCoalesce:
		return "COALESCE"
	case FuncNullif:
		return "NULLIF"
	case FuncConcat:
		return "CONCAT"
	case FuncUpper:
		return "UPPER"
	case FuncLower:
		return "LOWER"
	case FuncTrim:
		return "TRIM"
	case FuncLength:
		return "LENGTH"
	case FuncSubstring:
		return "SUBSTRING"
	case FuncAbs:
		return "ABS"
	case FuncRound:
		return "ROUND"
	case FuncFloor:
		return "FLOOR"
	case FuncCeil:
		return "CEIL"
	default:
		return "UNKNOWN"
	}
}

// MinArgs returns the minimum number of arguments for the function
func (f Function) MinArgs() int {
	switch f {
	case FuncCoalesce:
		return 1
	case FuncNullif:
		return 2
	case FuncConcat:
		return 2
	case FuncUpper, FuncLower, FuncTrim, FuncLength, FuncAbs, FuncFloor, FuncCeil:
		return 1
	case FuncSubstring:
		return 2 // SUBSTRING(str, start) or SUBSTRING(str, start, len)
	case FuncRound:
		return 1 // ROUND(num) or ROUND(num, decimals)
	default:
		return 0
	}
}

// MaxArgs returns the maximum number of arguments for the function (-1 for unlimited)
func (f Function) MaxArgs() int {
	switch f {
	case FuncCoalesce, FuncConcat:
		return -1 // unlimited
	case FuncNullif:
		return 2
	case FuncUpper, FuncLower, FuncTrim, FuncLength, FuncAbs, FuncFloor, FuncCeil:
		return 1
	case FuncSubstring:
		return 3
	case FuncRound:
		return 2
	default:
		return 0
	}
}

// ResultType returns the result type of the function based on argument types
func (f Function) ResultType(argTypes []FieldType) FieldType {
	switch f {
	case FuncCoalesce:
		// Returns the type of the first non-null argument
		for _, t := range argTypes {
			if t != FieldTypeNull {
				return t
			}
		}
		return FieldTypeNull
	case FuncNullif:
		if len(argTypes) > 0 {
			return argTypes[0]
		}
		return FieldTypeUnknown
	case FuncConcat, FuncUpper, FuncLower, FuncTrim, FuncSubstring:
		return FieldTypeString
	case FuncLength:
		return FieldTypeInteger
	case FuncAbs, FuncFloor, FuncCeil:
		if len(argTypes) > 0 {
			return argTypes[0]
		}
		return FieldTypeFloat
	case FuncRound:
		return FieldTypeFloat
	default:
		return FieldTypeUnknown
	}
}
