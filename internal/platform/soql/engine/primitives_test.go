package engine

import (
	"testing"
)

func TestFieldTypeString(t *testing.T) {
	tests := []struct {
		ft   FieldType
		want string
	}{
		{FieldTypeUnknown, "unknown"},
		{FieldTypeNull, "null"},
		{FieldTypeString, "string"},
		{FieldTypeInteger, "integer"},
		{FieldTypeFloat, "float"},
		{FieldTypeBoolean, "boolean"},
		{FieldTypeDate, "date"},
		{FieldTypeDateTime, "datetime"},
		{FieldTypeID, "id"},
		{FieldType(100), "unknown"}, // unknown base type
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ft.String(); got != tt.want {
				t.Errorf("FieldType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFieldTypeArray(t *testing.T) {
	t.Run("array string", func(t *testing.T) {
		ft := FieldTypeString | FieldTypeArray
		if got := ft.String(); got != "string[]" {
			t.Errorf("String() = %q, want %q", got, "string[]")
		}
	})

	t.Run("array integer", func(t *testing.T) {
		ft := FieldTypeInteger | FieldTypeArray
		if got := ft.String(); got != "integer[]" {
			t.Errorf("String() = %q, want %q", got, "integer[]")
		}
	})
}

func TestFieldTypeBase(t *testing.T) {
	tests := []struct {
		ft   FieldType
		want FieldType
	}{
		{FieldTypeString, FieldTypeString},
		{FieldTypeString | FieldTypeArray, FieldTypeString},
		{FieldTypeInteger | FieldTypeArray, FieldTypeInteger},
	}

	for _, tt := range tests {
		t.Run(tt.ft.String(), func(t *testing.T) {
			if got := tt.ft.Base(); got != tt.want {
				t.Errorf("Base() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldTypeIsArray(t *testing.T) {
	tests := []struct {
		ft   FieldType
		want bool
	}{
		{FieldTypeString, false},
		{FieldTypeString | FieldTypeArray, true},
		{FieldTypeInteger, false},
		{FieldTypeInteger | FieldTypeArray, true},
	}

	for _, tt := range tests {
		t.Run(tt.ft.String(), func(t *testing.T) {
			if got := tt.ft.IsArray(); got != tt.want {
				t.Errorf("IsArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatorCapture(t *testing.T) {
	tests := []struct {
		input string
		want  Operator
	}{
		{"*", OpMul},
		{"/", OpDiv},
		{"%", OpMod},
		{"+", OpAdd},
		{"-", OpSub},
		{"||", OpConcat},
		{"=", OpEQ},
		{"==", OpEQ},
		{"!=", OpNE},
		{"<>", OpNE},
		{">", OpGT},
		{"<", OpLT},
		{">=", OpGE},
		{"<=", OpLE},
		{"AND", OpAnd},
		{"and", OpAnd},
		{"OR", OpOr},
		{"or", OpOr},
		{"NOT", OpNot},
		{"LIKE", OpLike},
		{"IN", OpIn},
		{"IS", OpIs},
		{" = ", OpEQ},      // with whitespace
		{"  AND  ", OpAnd}, // with more whitespace
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var op Operator
			err := op.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if op != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, op, tt.want)
			}
		})
	}
}

func TestOperatorString(t *testing.T) {
	tests := []struct {
		op   Operator
		want string
	}{
		{OpMul, "*"},
		{OpDiv, "/"},
		{OpMod, "%"},
		{OpAdd, "+"},
		{OpSub, "-"},
		{OpConcat, "||"},
		{OpEQ, "="},
		{OpNE, "!="},
		{OpGT, ">"},
		{OpLT, "<"},
		{OpGE, ">="},
		{OpLE, "<="},
		{OpAnd, "AND"},
		{OpOr, "OR"},
		{OpNot, "NOT"},
		{OpLike, "LIKE"},
		{OpIn, "IN"},
		{OpIs, "IS"},
		{Operator(99), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.op.String(); got != tt.want {
				t.Errorf("Operator.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAggregateCapture(t *testing.T) {
	tests := []struct {
		input string
		want  Aggregate
	}{
		{"COUNT", AggregateCount},
		{"count", AggregateCount},
		{"COUNT_DISTINCT", AggregateCountDistinct},
		{"SUM", AggregateSum},
		{"AVG", AggregateAvg},
		{"MIN", AggregateMin},
		{"MAX", AggregateMax},
		{"  SUM  ", AggregateSum},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var agg Aggregate
			err := agg.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if agg != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, agg, tt.want)
			}
		})
	}
}

func TestAggregateString(t *testing.T) {
	tests := []struct {
		agg  Aggregate
		want string
	}{
		{AggregateCount, "COUNT"},
		{AggregateCountDistinct, "COUNT_DISTINCT"},
		{AggregateSum, "SUM"},
		{AggregateAvg, "AVG"},
		{AggregateMin, "MIN"},
		{AggregateMax, "MAX"},
		{Aggregate(99), "COUNT"}, // default
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.agg.String(); got != tt.want {
				t.Errorf("Aggregate.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirectionCapture(t *testing.T) {
	tests := []struct {
		input string
		want  Direction
	}{
		{"ASC", DirAsc},
		{"asc", DirAsc},
		{"DESC", DirDesc},
		{"desc", DirDesc},
		{"  ASC  ", DirAsc},
		{"anything", DirAsc}, // default to ASC
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var dir Direction
			err := dir.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if dir != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, dir, tt.want)
			}
		})
	}
}

func TestDirectionString(t *testing.T) {
	tests := []struct {
		dir  Direction
		want string
	}{
		{DirAsc, "ASC"},
		{DirDesc, "DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dir.String(); got != tt.want {
				t.Errorf("Direction.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNullsOrderCapture(t *testing.T) {
	tests := []struct {
		input string
		want  NullsOrder
	}{
		{"FIRST", NullsFirst},
		{"first", NullsFirst},
		{"LAST", NullsLast},
		{"last", NullsLast},
		{"other", NullsDefault},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var no NullsOrder
			err := no.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if no != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, no, tt.want)
			}
		})
	}
}

func TestNullsOrderString(t *testing.T) {
	tests := []struct {
		no   NullsOrder
		want string
	}{
		{NullsFirst, "NULLS FIRST"},
		{NullsLast, "NULLS LAST"},
		{NullsDefault, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.no.String(); got != tt.want {
				t.Errorf("NullsOrder.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStaticDateLiteralCapture(t *testing.T) {
	tests := []struct {
		input string
		want  StaticDateLiteral
	}{
		{"TODAY", DateToday},
		{"today", DateToday},
		{"YESTERDAY", DateYesterday},
		{"TOMORROW", DateTomorrow},
		{"THIS_WEEK", DateThisWeek},
		{"LAST_WEEK", DateLastWeek},
		{"NEXT_WEEK", DateNextWeek},
		{"THIS_MONTH", DateThisMonth},
		{"LAST_MONTH", DateLastMonth},
		{"NEXT_MONTH", DateNextMonth},
		{"THIS_QUARTER", DateThisQuarter},
		{"LAST_QUARTER", DateLastQuarter},
		{"NEXT_QUARTER", DateNextQuarter},
		{"THIS_YEAR", DateThisYear},
		{"LAST_YEAR", DateLastYear},
		{"NEXT_YEAR", DateNextYear},
		{"LAST_90_DAYS", DateLast90Days},
		{"NEXT_90_DAYS", DateNext90Days},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var dl StaticDateLiteral
			err := dl.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if dl != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, dl, tt.want)
			}
		})
	}
}

func TestStaticDateLiteralString(t *testing.T) {
	tests := []struct {
		dl   StaticDateLiteral
		want string
	}{
		{DateToday, "TODAY"},
		{DateYesterday, "YESTERDAY"},
		{DateTomorrow, "TOMORROW"},
		{DateThisWeek, "THIS_WEEK"},
		{DateLastWeek, "LAST_WEEK"},
		{DateNextWeek, "NEXT_WEEK"},
		{DateThisMonth, "THIS_MONTH"},
		{DateLastMonth, "LAST_MONTH"},
		{DateNextMonth, "NEXT_MONTH"},
		{DateThisQuarter, "THIS_QUARTER"},
		{DateLastQuarter, "LAST_QUARTER"},
		{DateNextQuarter, "NEXT_QUARTER"},
		{DateThisYear, "THIS_YEAR"},
		{DateLastYear, "LAST_YEAR"},
		{DateNextYear, "NEXT_YEAR"},
		{DateLast90Days, "LAST_90_DAYS"},
		{DateNext90Days, "NEXT_90_DAYS"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dl.String(); got != tt.want {
				t.Errorf("StaticDateLiteral.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDynamicDateTypeString(t *testing.T) {
	tests := []struct {
		dt   DynamicDateType
		want string
	}{
		{DynamicLastNDays, "LAST_N_DAYS"},
		{DynamicNextNDays, "NEXT_N_DAYS"},
		{DynamicLastNWeeks, "LAST_N_WEEKS"},
		{DynamicNextNWeeks, "NEXT_N_WEEKS"},
		{DynamicLastNMonths, "LAST_N_MONTHS"},
		{DynamicNextNMonths, "NEXT_N_MONTHS"},
		{DynamicLastNQuarters, "LAST_N_QUARTERS"},
		{DynamicNextNQuarters, "NEXT_N_QUARTERS"},
		{DynamicLastNYears, "LAST_N_YEARS"},
		{DynamicNextNYears, "NEXT_N_YEARS"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.dt.String(); got != tt.want {
				t.Errorf("DynamicDateType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDynamicDateLiteralCapture(t *testing.T) {
	tests := []struct {
		input    string
		wantType DynamicDateType
		wantN    int
	}{
		{"LAST_N_DAYS:30", DynamicLastNDays, 30},
		{"NEXT_N_DAYS:7", DynamicNextNDays, 7},
		{"LAST_N_WEEKS:4", DynamicLastNWeeks, 4},
		{"NEXT_N_WEEKS:2", DynamicNextNWeeks, 2},
		{"LAST_N_MONTHS:6", DynamicLastNMonths, 6},
		{"NEXT_N_MONTHS:3", DynamicNextNMonths, 3},
		{"LAST_N_QUARTERS:2", DynamicLastNQuarters, 2},
		{"NEXT_N_QUARTERS:1", DynamicNextNQuarters, 1},
		{"LAST_N_YEARS:5", DynamicLastNYears, 5},
		{"NEXT_N_YEARS:2", DynamicNextNYears, 2},
		{"last_n_days:15", DynamicLastNDays, 15}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var dl DynamicDateLiteral
			err := dl.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if dl.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", dl.Type, tt.wantType)
			}
			if dl.N != tt.wantN {
				t.Errorf("N = %d, want %d", dl.N, tt.wantN)
			}
		})
	}
}

func TestDynamicDateLiteralCaptureInvalid(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"LAST_N_DAYS:abc"}, // non-numeric
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var dl DynamicDateLiteral
			err := dl.Capture([]string{tt.input})
			if err == nil {
				t.Error("expected error for invalid input")
			}
		})
	}
}

func TestDynamicDateLiteralString(t *testing.T) {
	dl := DynamicDateLiteral{Type: DynamicLastNDays, N: 30}
	want := "LAST_N_DAYS:30"
	if got := dl.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBooleanCapture(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"TRUE", true},
		{"true", true},
		{"True", true},
		{"FALSE", false},
		{"false", false},
		{"False", false},
		{"other", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var b Boolean
			err := b.Capture([]string{tt.input})
			if err != nil {
				t.Fatalf("Capture() error = %v", err)
			}
			if bool(b) != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, b, tt.want)
			}
		})
	}
}

func TestDateCapture(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-01-15", false},
		{"2024-12-31", false},
		{"2024-02-29", false}, // leap year
		{"invalid", true},
		{"2024/01/15", true}, // wrong format
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var d Date
			err := d.Capture([]string{tt.input})
			if (err != nil) != tt.wantErr {
				t.Errorf("Capture(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				expected := tt.input
				if d.Format("2006-01-02") != expected {
					t.Errorf("Date = %v, want %v", d.Format("2006-01-02"), expected)
				}
			}
		})
	}
}

func TestDateTimeCapture(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"2024-01-15T10:30:00Z", false},
		{"2024-12-31T23:59:59Z", false},
		{"2024-01-15T10:30:00+05:00", false},
		{"invalid", true},
		{"2024-01-15", true}, // date only
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var dt DateTime
			err := dt.Capture([]string{tt.input})
			if (err != nil) != tt.wantErr {
				t.Errorf("Capture(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestFunctionCapture(t *testing.T) {
	tests := []struct {
		input   string
		want    Function
		wantErr bool
	}{
		{"COALESCE", FuncCoalesce, false},
		{"coalesce", FuncCoalesce, false},
		{"NULLIF", FuncNullif, false},
		{"CONCAT", FuncConcat, false},
		{"UPPER", FuncUpper, false},
		{"LOWER", FuncLower, false},
		{"TRIM", FuncTrim, false},
		{"LENGTH", FuncLength, false},
		{"LEN", FuncLength, false},
		{"SUBSTRING", FuncSubstring, false},
		{"SUBSTR", FuncSubstring, false},
		{"ABS", FuncAbs, false},
		{"ROUND", FuncRound, false},
		{"FLOOR", FuncFloor, false},
		{"CEIL", FuncCeil, false},
		{"CEILING", FuncCeil, false},
		{"UNKNOWN_FUNC", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var f Function
			err := f.Capture([]string{tt.input})
			if (err != nil) != tt.wantErr {
				t.Errorf("Capture(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && f != tt.want {
				t.Errorf("Capture(%q) = %v, want %v", tt.input, f, tt.want)
			}
		})
	}
}

func TestFunctionString(t *testing.T) {
	tests := []struct {
		f    Function
		want string
	}{
		{FuncCoalesce, "COALESCE"},
		{FuncNullif, "NULLIF"},
		{FuncConcat, "CONCAT"},
		{FuncUpper, "UPPER"},
		{FuncLower, "LOWER"},
		{FuncTrim, "TRIM"},
		{FuncLength, "LENGTH"},
		{FuncSubstring, "SUBSTRING"},
		{FuncAbs, "ABS"},
		{FuncRound, "ROUND"},
		{FuncFloor, "FLOOR"},
		{FuncCeil, "CEIL"},
		{Function(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("Function.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFunctionMinArgs(t *testing.T) {
	tests := []struct {
		f    Function
		want int
	}{
		{FuncCoalesce, 1},
		{FuncNullif, 2},
		{FuncConcat, 2},
		{FuncUpper, 1},
		{FuncLower, 1},
		{FuncTrim, 1},
		{FuncLength, 1},
		{FuncSubstring, 2},
		{FuncAbs, 1},
		{FuncRound, 1},
		{FuncFloor, 1},
		{FuncCeil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.f.String(), func(t *testing.T) {
			if got := tt.f.MinArgs(); got != tt.want {
				t.Errorf("MinArgs() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFunctionMaxArgs(t *testing.T) {
	tests := []struct {
		f    Function
		want int
	}{
		{FuncCoalesce, -1}, // unlimited
		{FuncConcat, -1},   // unlimited
		{FuncNullif, 2},
		{FuncUpper, 1},
		{FuncLower, 1},
		{FuncTrim, 1},
		{FuncLength, 1},
		{FuncSubstring, 3},
		{FuncAbs, 1},
		{FuncRound, 2},
		{FuncFloor, 1},
		{FuncCeil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.f.String(), func(t *testing.T) {
			if got := tt.f.MaxArgs(); got != tt.want {
				t.Errorf("MaxArgs() = %d, want %d", got, tt.want)
			}
		})
	}
}
