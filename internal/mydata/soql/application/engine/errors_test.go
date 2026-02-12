package engine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
)

func TestErrorTypeString(t *testing.T) {
	tests := []struct {
		errType ErrorType
		want    string
	}{
		{ErrTypeParse, "ParseError"},
		{ErrTypeValidation, "ValidationError"},
		{ErrTypeAccess, "AccessError"},
		{ErrTypeLimit, "LimitError"},
		{ErrTypeExecution, "ExecutionError"},
		{ErrorType(99), "UnknownError"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.errType.String(); got != tt.want {
				t.Errorf("ErrorType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPositionString(t *testing.T) {
	tests := []struct {
		pos  Position
		want string
	}{
		{Position{Line: 1, Column: 1}, "line 1, column 1"},
		{Position{Line: 5, Column: 10}, "line 5, column 10"},
		{Position{Line: 0, Column: 0}, "line 0, column 0"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pos.String(); got != tt.want {
				t.Errorf("Position.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPosFromLexer(t *testing.T) {
	lp := lexer.Position{
		Line:   5,
		Column: 10,
		Offset: 50,
	}

	pos := PosFromLexer(lp)

	if pos.Line != 5 {
		t.Errorf("Line = %d, want 5", pos.Line)
	}
	if pos.Column != 10 {
		t.Errorf("Column = %d, want 10", pos.Column)
	}
	if pos.Offset != 50 {
		t.Errorf("Offset = %d, want 50", pos.Offset)
	}
}

func TestBaseErrorError(t *testing.T) {
	t.Run("with position", func(t *testing.T) {
		err := &BaseError{
			Type:    ErrTypeParse,
			Message: "unexpected token",
			Pos:     Position{Line: 1, Column: 10},
		}

		got := err.Error()
		want := "ParseError at line 1, column 10: unexpected token"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("without position", func(t *testing.T) {
		err := &BaseError{
			Type:    ErrTypeValidation,
			Message: "unknown object",
		}

		got := err.Error()
		want := "ValidationError: unknown object"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestBaseErrorUnwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &BaseError{
		Type:    ErrTypeExecution,
		Message: "execution failed",
		Cause:   cause,
	}

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestNewParseError(t *testing.T) {
	pos := Position{Line: 1, Column: 5}
	err := NewParseError(pos, "identifier", "number")

	if err.Type != ErrTypeParse {
		t.Errorf("Type = %v, want ErrTypeParse", err.Type)
	}
	if err.Pos != pos {
		t.Errorf("Pos = %v, want %v", err.Pos, pos)
	}
	if err.Expected != "identifier" {
		t.Errorf("Expected = %q, want %q", err.Expected, "identifier")
	}
	if err.Got != "number" {
		t.Errorf("Got = %q, want %q", err.Got, "number")
	}
}

func TestNewParseErrorFromParticiple(t *testing.T) {
	originalErr := fmt.Errorf("1:10: unexpected token 'FOO'")
	err := NewParseErrorFromParticiple(originalErr)

	if err.Type != ErrTypeParse {
		t.Errorf("Type = %v, want ErrTypeParse", err.Type)
	}
	if err.Pos.Line != 1 {
		t.Errorf("Pos.Line = %d, want 1", err.Pos.Line)
	}
	if err.Pos.Column != 10 {
		t.Errorf("Pos.Column = %d, want 10", err.Pos.Column)
	}
	if err.Cause != originalErr {
		t.Errorf("Cause should be the original error")
	}
}

func TestParsePositionFromMessage(t *testing.T) {
	tests := []struct {
		msg      string
		wantLine int
		wantCol  int
	}{
		{"1:10: message", 1, 10},
		{"5:25: another message", 5, 25},
		{"no position here", 0, 0},
		{"", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			pos := parsePositionFromMessage(tt.msg)
			if pos.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", pos.Line, tt.wantLine)
			}
			if pos.Column != tt.wantCol {
				t.Errorf("Column = %d, want %d", pos.Column, tt.wantCol)
			}
		})
	}
}

func TestValidationErrorCodeString(t *testing.T) {
	tests := []struct {
		code ValidationErrorCode
		want string
	}{
		{ErrCodeUnknownObject, "UnknownObject"},
		{ErrCodeUnknownField, "UnknownField"},
		{ErrCodeUnknownLookup, "UnknownLookup"},
		{ErrCodeUnknownRelationship, "UnknownRelationship"},
		{ErrCodeTypeMismatch, "TypeMismatch"},
		{ErrCodeInvalidOperator, "InvalidOperator"},
		{ErrCodeInvalidAggregation, "InvalidAggregation"},
		{ErrCodeFieldNotFilterable, "FieldNotFilterable"},
		{ErrCodeFieldNotSortable, "FieldNotSortable"},
		{ErrCodeFieldNotGroupable, "FieldNotGroupable"},
		{ErrCodeFieldNotAggregatable, "FieldNotAggregatable"},
		{ErrCodeNestedSubqueryNotAllowed, "NestedSubqueryNotAllowed"},
		{ErrCodeTooManyLookupLevels, "TooManyLookupLevels"},
		{ErrCodeInvalidExpression, "InvalidExpression"},
		{ErrCodeMissingRequiredClause, "MissingRequiredClause"},
		{ErrCodeInvalidDateLiteral, "InvalidDateLiteral"},
		{ErrCodeInvalidPagination, "InvalidPagination"},
		{ErrCodeWhereSubquerySingleField, "WhereSubquerySingleField"},
		{ErrCodeWhereSubqueryAggregateField, "WhereSubqueryAggregateField"},
		{ValidationErrorCode(999), "UnknownValidationError"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.code.String(); got != tt.want {
				t.Errorf("ValidationErrorCode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError(ErrCodeUnknownObject, "object not found")

	if err.Type != ErrTypeValidation {
		t.Errorf("Type = %v, want ErrTypeValidation", err.Type)
	}
	if err.Code != ErrCodeUnknownObject {
		t.Errorf("Code = %v, want ErrCodeUnknownObject", err.Code)
	}
	if err.Message != "object not found" {
		t.Errorf("Message = %q, want %q", err.Message, "object not found")
	}
}

func TestNewValidationErrorWithPos(t *testing.T) {
	pos := Position{Line: 3, Column: 15}
	err := NewValidationErrorWithPos(ErrCodeUnknownField, pos, "field not found")

	if err.Pos != pos {
		t.Errorf("Pos = %v, want %v", err.Pos, pos)
	}
}

func TestUnknownObjectError(t *testing.T) {
	err := UnknownObjectError("MyObject")

	if err.Code != ErrCodeUnknownObject {
		t.Errorf("Code = %v, want ErrCodeUnknownObject", err.Code)
	}
	if err.Object != "MyObject" {
		t.Errorf("Object = %q, want %q", err.Object, "MyObject")
	}
}

func TestUnknownFieldError(t *testing.T) {
	err := UnknownFieldError("Account", "InvalidField")

	if err.Code != ErrCodeUnknownField {
		t.Errorf("Code = %v, want ErrCodeUnknownField", err.Code)
	}
	if err.Object != "Account" {
		t.Errorf("Object = %q, want %q", err.Object, "Account")
	}
	if err.Field != "InvalidField" {
		t.Errorf("Field = %q, want %q", err.Field, "InvalidField")
	}
}

func TestUnknownFieldErrorAt(t *testing.T) {
	pos := lexer.Position{Line: 2, Column: 10, Offset: 20}
	err := UnknownFieldErrorAt("Account", "InvalidField", pos)

	if err.Pos.Line != 2 {
		t.Errorf("Pos.Line = %d, want 2", err.Pos.Line)
	}
}

func TestUnknownLookupError(t *testing.T) {
	err := UnknownLookupError("Contact", "InvalidLookup")

	if err.Code != ErrCodeUnknownLookup {
		t.Errorf("Code = %v, want ErrCodeUnknownLookup", err.Code)
	}
}

func TestUnknownRelationshipError(t *testing.T) {
	err := UnknownRelationshipError("Account", "InvalidRel")

	if err.Code != ErrCodeUnknownRelationship {
		t.Errorf("Code = %v, want ErrCodeUnknownRelationship", err.Code)
	}
}

func TestTypeMismatchError(t *testing.T) {
	err := TypeMismatchError(FieldTypeString, FieldTypeInteger)

	if err.Code != ErrCodeTypeMismatch {
		t.Errorf("Code = %v, want ErrCodeTypeMismatch", err.Code)
	}
}

func TestFieldNotFilterableError(t *testing.T) {
	err := FieldNotFilterableError("Account", "Description")

	if err.Code != ErrCodeFieldNotFilterable {
		t.Errorf("Code = %v, want ErrCodeFieldNotFilterable", err.Code)
	}
}

func TestFieldNotSortableError(t *testing.T) {
	err := FieldNotSortableError("Account", "LargeText")

	if err.Code != ErrCodeFieldNotSortable {
		t.Errorf("Code = %v, want ErrCodeFieldNotSortable", err.Code)
	}
}

func TestFieldNotGroupableError(t *testing.T) {
	err := FieldNotGroupableError("Account", "Blob")

	if err.Code != ErrCodeFieldNotGroupable {
		t.Errorf("Code = %v, want ErrCodeFieldNotGroupable", err.Code)
	}
}

func TestWhereSubquerySingleFieldError(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 50}
	err := WhereSubquerySingleFieldError(pos)

	if err.Code != ErrCodeWhereSubquerySingleField {
		t.Errorf("Code = %v, want ErrCodeWhereSubquerySingleField", err.Code)
	}
	if err.Pos.Line != 1 {
		t.Errorf("Pos.Line = %d, want 1", err.Pos.Line)
	}
}

func TestWhereSubqueryAggregateFieldError(t *testing.T) {
	pos := lexer.Position{Line: 1, Column: 60}
	err := WhereSubqueryAggregateFieldError(pos)

	if err.Code != ErrCodeWhereSubqueryAggregateField {
		t.Errorf("Code = %v, want ErrCodeWhereSubqueryAggregateField", err.Code)
	}
}

func TestNewAccessError(t *testing.T) {
	err := NewAccessError("Account")

	if err.Type != ErrTypeAccess {
		t.Errorf("Type = %v, want ErrTypeAccess", err.Type)
	}
	if err.Object != "Account" {
		t.Errorf("Object = %q, want %q", err.Object, "Account")
	}
	if err.Field != "" {
		t.Errorf("Field = %q, want empty", err.Field)
	}
}

func TestNewFieldAccessError(t *testing.T) {
	err := NewFieldAccessError("Account", "SSN")

	if err.Type != ErrTypeAccess {
		t.Errorf("Type = %v, want ErrTypeAccess", err.Type)
	}
	if err.Object != "Account" {
		t.Errorf("Object = %q, want %q", err.Object, "Account")
	}
	if err.Field != "SSN" {
		t.Errorf("Field = %q, want %q", err.Field, "SSN")
	}
}

func TestLimitTypeString(t *testing.T) {
	tests := []struct {
		limitType LimitType
		want      string
	}{
		{LimitTypeMaxFields, "MaxFields"},
		{LimitTypeMaxRecords, "MaxRecords"},
		{LimitTypeMaxSubqueries, "MaxSubqueries"},
		{LimitTypeMaxLookupDepth, "MaxLookupDepth"},
		{LimitTypeMaxSubqueryRecords, "MaxSubqueryRecords"},
		{LimitTypeMaxOffset, "MaxOffset"},
		{LimitTypeMaxQueryLength, "MaxQueryLength"},
		{LimitType(99), "UnknownLimit"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.limitType.String(); got != tt.want {
				t.Errorf("LimitType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewLimitError(t *testing.T) {
	err := NewLimitError(LimitTypeMaxRecords, 1000, 5000)

	if err.Type != ErrTypeLimit {
		t.Errorf("Type = %v, want ErrTypeLimit", err.Type)
	}
	if err.LimitType != LimitTypeMaxRecords {
		t.Errorf("LimitType = %v, want LimitTypeMaxRecords", err.LimitType)
	}
	if err.Limit != 1000 {
		t.Errorf("Limit = %d, want 1000", err.Limit)
	}
	if err.Actual != 5000 {
		t.Errorf("Actual = %d, want 5000", err.Actual)
	}
}

func TestNewExecutionError(t *testing.T) {
	sqlErr := errors.New("database error")
	err := NewExecutionError("query failed", sqlErr)

	if err.Type != ErrTypeExecution {
		t.Errorf("Type = %v, want ErrTypeExecution", err.Type)
	}
	if err.SQLError != sqlErr {
		t.Errorf("SQLError = %v, want %v", err.SQLError, sqlErr)
	}
	if err.SQL != "" {
		t.Errorf("SQL = %q, want empty", err.SQL)
	}
}

func TestNewExecutionErrorWithSQL(t *testing.T) {
	sqlErr := errors.New("database error")
	err := NewExecutionErrorWithSQL("query failed", "SELECT * FROM invalid", sqlErr)

	if err.SQL != "SELECT * FROM invalid" {
		t.Errorf("SQL = %q, want %q", err.SQL, "SELECT * FROM invalid")
	}
}

func TestIsParseError(t *testing.T) {
	parseErr := NewParseError(Position{}, "expected", "got")
	validErr := NewValidationError(ErrCodeUnknownObject, "test")
	plainErr := errors.New("plain error")

	if !IsParseError(parseErr) {
		t.Error("IsParseError should return true for ParseError")
	}
	if IsParseError(validErr) {
		t.Error("IsParseError should return false for ValidationError")
	}
	if IsParseError(plainErr) {
		t.Error("IsParseError should return false for plain error")
	}
}

func TestIsValidationError(t *testing.T) {
	parseErr := NewParseError(Position{}, "expected", "got")
	validErr := NewValidationError(ErrCodeUnknownObject, "test")

	if IsValidationError(parseErr) {
		t.Error("IsValidationError should return false for ParseError")
	}
	if !IsValidationError(validErr) {
		t.Error("IsValidationError should return true for ValidationError")
	}
}

func TestIsAccessError(t *testing.T) {
	accessErr := NewAccessError("Account")
	validErr := NewValidationError(ErrCodeUnknownObject, "test")

	if !IsAccessError(accessErr) {
		t.Error("IsAccessError should return true for AccessError")
	}
	if IsAccessError(validErr) {
		t.Error("IsAccessError should return false for ValidationError")
	}
}

func TestIsLimitError(t *testing.T) {
	limitErr := NewLimitError(LimitTypeMaxRecords, 100, 200)
	validErr := NewValidationError(ErrCodeUnknownObject, "test")

	if !IsLimitError(limitErr) {
		t.Error("IsLimitError should return true for LimitError")
	}
	if IsLimitError(validErr) {
		t.Error("IsLimitError should return false for ValidationError")
	}
}

func TestIsExecutionError(t *testing.T) {
	execErr := NewExecutionError("failed", errors.New("db error"))
	validErr := NewValidationError(ErrCodeUnknownObject, "test")

	if !IsExecutionError(execErr) {
		t.Error("IsExecutionError should return true for ExecutionError")
	}
	if IsExecutionError(validErr) {
		t.Error("IsExecutionError should return false for ValidationError")
	}
}

func TestErrorsAsUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	execErr := &ExecutionError{
		BaseError: BaseError{
			Type:    ErrTypeExecution,
			Message: "execution failed",
			Cause:   cause,
		},
	}

	// Test errors.Is works with Unwrap
	if !errors.Is(execErr, cause) {
		t.Error("errors.Is should match the cause")
	}

	// Test errors.As works for ExecutionError
	var extractedExecErr *ExecutionError
	if !errors.As(execErr, &extractedExecErr) {
		t.Error("errors.As should extract ExecutionError")
	}

	// Verify the extracted error has correct fields
	if extractedExecErr.Type != ErrTypeExecution {
		t.Errorf("Type = %v, want ErrTypeExecution", extractedExecErr.Type)
	}
}

func TestWrappedErrors(t *testing.T) {
	parseErr := NewParseError(Position{Line: 1, Column: 5}, "SELECT", "INSERT")
	wrappedErr := fmt.Errorf("query processing failed: %w", parseErr)

	if !IsParseError(wrappedErr) {
		t.Error("IsParseError should work with wrapped errors")
	}

	var pe *ParseError
	if !errors.As(wrappedErr, &pe) {
		t.Error("errors.As should extract ParseError from wrapped error")
	}
	if pe.Expected != "SELECT" {
		t.Errorf("Expected = %q, want %q", pe.Expected, "SELECT")
	}
}
