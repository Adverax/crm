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
	FieldTypeObject // JSON object type for TYPEOF expressions
)

const (
	FieldTypeArray FieldType = 256
)

func (t FieldType) String() string {
	base := t & 0xFF
	var s string
	switch base {
	case FieldTypeUnknown:
		s = "unknown"
	case FieldTypeNull:
		s = "null"
	case FieldTypeString:
		s = "string"
	case FieldTypeInteger:
		s = "integer"
	case FieldTypeFloat:
		s = "float"
	case FieldTypeBoolean:
		s = "boolean"
	case FieldTypeDate:
		s = "date"
	case FieldTypeDateTime:
		s = "datetime"
	case FieldTypeID:
		s = "id"
	case FieldTypeObject:
		s = "object"
	default:
		s = "unknown"
	}
	if t.IsArray() {
		s += "[]"
	}
	return s
}

func (t FieldType) Base() FieldType {
	return t & 0xFF
}

func (t FieldType) IsArray() bool {
	return t&FieldTypeArray != 0
}

// Operator represents comparison and arithmetic operators
type Operator int

const (
	OpMul Operator = iota
	OpDiv
	OpMod
	OpAdd
	OpSub
	OpConcat // || string concatenation
	OpEQ
	OpNE
	OpGT
	OpLT
	OpGE
	OpLE
	OpAnd
	OpOr
	OpNot
	OpLike
	OpIn
	OpIs
)

func (op *Operator) Capture(s []string) error {
	val := strings.TrimSpace(s[0])
	switch val {
	case "*":
		*op = OpMul
	case "/":
		*op = OpDiv
	case "%":
		*op = OpMod
	case "+":
		*op = OpAdd
	case "-":
		*op = OpSub
	case "||":
		*op = OpConcat
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
		// Handle keywords (case-insensitive)
		upper := strings.ToUpper(val)
		switch upper {
		case "AND":
			*op = OpAnd
		case "OR":
			*op = OpOr
		case "NOT":
			*op = OpNot
		case "LIKE":
			*op = OpLike
		case "IN":
			*op = OpIn
		case "IS":
			*op = OpIs
		default:
			*op = OpEQ
		}
	}
	return nil
}

func (op Operator) String() string {
	switch op {
	case OpMul:
		return "*"
	case OpDiv:
		return "/"
	case OpMod:
		return "%"
	case OpAdd:
		return "+"
	case OpSub:
		return "-"
	case OpConcat:
		return "||"
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
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpNot:
		return "NOT"
	case OpLike:
		return "LIKE"
	case OpIn:
		return "IN"
	case OpIs:
		return "IS"
	default:
		return "?"
	}
}

// Aggregate represents aggregate functions
type Aggregate int

const (
	AggregateCount Aggregate = iota
	AggregateCountDistinct
	AggregateSum
	AggregateAvg
	AggregateMin
	AggregateMax
)

func (a *Aggregate) Capture(s []string) error {
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	switch val {
	case "COUNT":
		*a = AggregateCount
	case "COUNT_DISTINCT":
		*a = AggregateCountDistinct
	case "SUM":
		*a = AggregateSum
	case "AVG":
		*a = AggregateAvg
	case "MIN":
		*a = AggregateMin
	case "MAX":
		*a = AggregateMax
	default:
		*a = AggregateCount
	}
	return nil
}

func (a Aggregate) String() string {
	switch a {
	case AggregateCount:
		return "COUNT"
	case AggregateCountDistinct:
		return "COUNT_DISTINCT"
	case AggregateSum:
		return "SUM"
	case AggregateAvg:
		return "AVG"
	case AggregateMin:
		return "MIN"
	case AggregateMax:
		return "MAX"
	default:
		return "COUNT"
	}
}

// Direction represents ORDER BY direction
type Direction int

const (
	DirAsc Direction = iota
	DirDesc
)

func (d *Direction) Capture(s []string) error {
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	if val == "DESC" {
		*d = DirDesc
	} else {
		*d = DirAsc
	}
	return nil
}

func (d Direction) String() string {
	if d == DirDesc {
		return "DESC"
	}
	return "ASC"
}

// NullsOrder represents NULLS FIRST/LAST in ORDER BY
type NullsOrder int

const (
	NullsDefault NullsOrder = iota
	NullsFirst
	NullsLast
)

func (n *NullsOrder) Capture(s []string) error {
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	switch val {
	case "FIRST":
		*n = NullsFirst
	case "LAST":
		*n = NullsLast
	default:
		*n = NullsDefault
	}
	return nil
}

func (n NullsOrder) String() string {
	switch n {
	case NullsFirst:
		return "NULLS FIRST"
	case NullsLast:
		return "NULLS LAST"
	default:
		return ""
	}
}

// StaticDateLiteral represents static date literals like TODAY, YESTERDAY
type StaticDateLiteral int

const (
	DateToday StaticDateLiteral = iota
	DateYesterday
	DateTomorrow
	DateThisWeek
	DateLastWeek
	DateNextWeek
	DateThisMonth
	DateLastMonth
	DateNextMonth
	DateThisQuarter
	DateLastQuarter
	DateNextQuarter
	DateThisYear
	DateLastYear
	DateNextYear
	DateLast90Days
	DateNext90Days
	// Fiscal periods
	DateThisFiscalQuarter
	DateLastFiscalQuarter
	DateNextFiscalQuarter
	DateThisFiscalYear
	DateLastFiscalYear
	DateNextFiscalYear
)

func (d *StaticDateLiteral) Capture(s []string) error {
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	switch val {
	case "TODAY":
		*d = DateToday
	case "YESTERDAY":
		*d = DateYesterday
	case "TOMORROW":
		*d = DateTomorrow
	case "THIS_WEEK":
		*d = DateThisWeek
	case "LAST_WEEK":
		*d = DateLastWeek
	case "NEXT_WEEK":
		*d = DateNextWeek
	case "THIS_MONTH":
		*d = DateThisMonth
	case "LAST_MONTH":
		*d = DateLastMonth
	case "NEXT_MONTH":
		*d = DateNextMonth
	case "THIS_QUARTER":
		*d = DateThisQuarter
	case "LAST_QUARTER":
		*d = DateLastQuarter
	case "NEXT_QUARTER":
		*d = DateNextQuarter
	case "THIS_YEAR":
		*d = DateThisYear
	case "LAST_YEAR":
		*d = DateLastYear
	case "NEXT_YEAR":
		*d = DateNextYear
	case "LAST_90_DAYS":
		*d = DateLast90Days
	case "NEXT_90_DAYS":
		*d = DateNext90Days
	case "THIS_FISCAL_QUARTER":
		*d = DateThisFiscalQuarter
	case "LAST_FISCAL_QUARTER":
		*d = DateLastFiscalQuarter
	case "NEXT_FISCAL_QUARTER":
		*d = DateNextFiscalQuarter
	case "THIS_FISCAL_YEAR":
		*d = DateThisFiscalYear
	case "LAST_FISCAL_YEAR":
		*d = DateLastFiscalYear
	case "NEXT_FISCAL_YEAR":
		*d = DateNextFiscalYear
	default:
		*d = DateToday
	}
	return nil
}

func (d StaticDateLiteral) String() string {
	switch d {
	case DateToday:
		return "TODAY"
	case DateYesterday:
		return "YESTERDAY"
	case DateTomorrow:
		return "TOMORROW"
	case DateThisWeek:
		return "THIS_WEEK"
	case DateLastWeek:
		return "LAST_WEEK"
	case DateNextWeek:
		return "NEXT_WEEK"
	case DateThisMonth:
		return "THIS_MONTH"
	case DateLastMonth:
		return "LAST_MONTH"
	case DateNextMonth:
		return "NEXT_MONTH"
	case DateThisQuarter:
		return "THIS_QUARTER"
	case DateLastQuarter:
		return "LAST_QUARTER"
	case DateNextQuarter:
		return "NEXT_QUARTER"
	case DateThisYear:
		return "THIS_YEAR"
	case DateLastYear:
		return "LAST_YEAR"
	case DateNextYear:
		return "NEXT_YEAR"
	case DateLast90Days:
		return "LAST_90_DAYS"
	case DateNext90Days:
		return "NEXT_90_DAYS"
	case DateThisFiscalQuarter:
		return "THIS_FISCAL_QUARTER"
	case DateLastFiscalQuarter:
		return "LAST_FISCAL_QUARTER"
	case DateNextFiscalQuarter:
		return "NEXT_FISCAL_QUARTER"
	case DateThisFiscalYear:
		return "THIS_FISCAL_YEAR"
	case DateLastFiscalYear:
		return "LAST_FISCAL_YEAR"
	case DateNextFiscalYear:
		return "NEXT_FISCAL_YEAR"
	default:
		return "TODAY"
	}
}

// DynamicDateType represents dynamic date literal types like LAST_N_DAYS
type DynamicDateType int

const (
	DynamicLastNDays DynamicDateType = iota
	DynamicNextNDays
	DynamicLastNWeeks
	DynamicNextNWeeks
	DynamicLastNMonths
	DynamicNextNMonths
	DynamicLastNQuarters
	DynamicNextNQuarters
	DynamicLastNYears
	DynamicNextNYears
	// Fiscal dynamic periods
	DynamicLastNFiscalQuarters
	DynamicNextNFiscalQuarters
	DynamicLastNFiscalYears
	DynamicNextNFiscalYears
)

func (d DynamicDateType) String() string {
	switch d {
	case DynamicLastNDays:
		return "LAST_N_DAYS"
	case DynamicNextNDays:
		return "NEXT_N_DAYS"
	case DynamicLastNWeeks:
		return "LAST_N_WEEKS"
	case DynamicNextNWeeks:
		return "NEXT_N_WEEKS"
	case DynamicLastNMonths:
		return "LAST_N_MONTHS"
	case DynamicNextNMonths:
		return "NEXT_N_MONTHS"
	case DynamicLastNQuarters:
		return "LAST_N_QUARTERS"
	case DynamicNextNQuarters:
		return "NEXT_N_QUARTERS"
	case DynamicLastNYears:
		return "LAST_N_YEARS"
	case DynamicNextNYears:
		return "NEXT_N_YEARS"
	case DynamicLastNFiscalQuarters:
		return "LAST_N_FISCAL_QUARTERS"
	case DynamicNextNFiscalQuarters:
		return "NEXT_N_FISCAL_QUARTERS"
	case DynamicLastNFiscalYears:
		return "LAST_N_FISCAL_YEARS"
	case DynamicNextNFiscalYears:
		return "NEXT_N_FISCAL_YEARS"
	default:
		return "LAST_N_DAYS"
	}
}

// DynamicDateLiteral represents a dynamic date literal like LAST_N_DAYS:30
type DynamicDateLiteral struct {
	Type DynamicDateType
	N    int
}

func (d *DynamicDateLiteral) Capture(s []string) error {
	// Input format: "LAST_N_DAYS:30"
	val := strings.ToUpper(strings.TrimSpace(s[0]))
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil
	}

	switch parts[0] {
	case "LAST_N_DAYS":
		d.Type = DynamicLastNDays
	case "NEXT_N_DAYS":
		d.Type = DynamicNextNDays
	case "LAST_N_WEEKS":
		d.Type = DynamicLastNWeeks
	case "NEXT_N_WEEKS":
		d.Type = DynamicNextNWeeks
	case "LAST_N_MONTHS":
		d.Type = DynamicLastNMonths
	case "NEXT_N_MONTHS":
		d.Type = DynamicNextNMonths
	case "LAST_N_QUARTERS":
		d.Type = DynamicLastNQuarters
	case "NEXT_N_QUARTERS":
		d.Type = DynamicNextNQuarters
	case "LAST_N_YEARS":
		d.Type = DynamicLastNYears
	case "NEXT_N_YEARS":
		d.Type = DynamicNextNYears
	case "LAST_N_FISCAL_QUARTERS":
		d.Type = DynamicLastNFiscalQuarters
	case "NEXT_N_FISCAL_QUARTERS":
		d.Type = DynamicNextNFiscalQuarters
	case "LAST_N_FISCAL_YEARS":
		d.Type = DynamicLastNFiscalYears
	case "NEXT_N_FISCAL_YEARS":
		d.Type = DynamicNextNFiscalYears
	}

	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	d.N = n
	return nil
}

func (d DynamicDateLiteral) String() string {
	return d.Type.String() + ":" + strconv.Itoa(d.N)
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

// Function represents built-in functions
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
