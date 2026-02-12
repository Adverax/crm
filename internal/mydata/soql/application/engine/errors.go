package engine

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

// ErrorType categorizes SOQL errors.
type ErrorType int

const (
	ErrTypeParse      ErrorType = iota // Syntax error during parsing
	ErrTypeValidation                  // Semantic error during validation
	ErrTypeAccess                      // Access denied error
	ErrTypeLimit                       // Limit exceeded error
	ErrTypeExecution                   // Runtime error during execution
)

func (t ErrorType) String() string {
	switch t {
	case ErrTypeParse:
		return "ParseError"
	case ErrTypeValidation:
		return "ValidationError"
	case ErrTypeAccess:
		return "AccessError"
	case ErrTypeLimit:
		return "LimitError"
	case ErrTypeExecution:
		return "ExecutionError"
	default:
		return "UnknownError"
	}
}

// Position represents a location in the SOQL query string.
type Position struct {
	Line   int // 1-based line number
	Column int // 1-based column number
	Offset int // 0-based byte offset
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

// PosFromLexer converts a lexer.Position to our Position type.
func PosFromLexer(lp lexer.Position) Position {
	return Position{
		Line:   lp.Line,
		Column: lp.Column,
		Offset: lp.Offset,
	}
}

// BaseError is the base error type for all SOQL errors.
type BaseError struct {
	Type    ErrorType
	Message string
	Pos     Position
	Cause   error
}

func (e *BaseError) Error() string {
	if e.Pos.Line > 0 {
		return fmt.Sprintf("%s at %s: %s", e.Type, e.Pos, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *BaseError) Unwrap() error {
	return e.Cause
}

// ParseError represents a syntax error during parsing.
type ParseError struct {
	BaseError
	Expected string // What was expected
	Got      string // What was actually found
}

// NewParseError creates a new ParseError.
func NewParseError(pos Position, expected, got string) *ParseError {
	return &ParseError{
		BaseError: BaseError{
			Type:    ErrTypeParse,
			Message: fmt.Sprintf("expected %s, got %s", expected, got),
			Pos:     pos,
		},
		Expected: expected,
		Got:      got,
	}
}

// PositionError is an interface for errors that have position information.
type PositionError interface {
	Position() Position
}

// NewParseErrorFromParticiple creates a ParseError from a participle error.
func NewParseErrorFromParticiple(err error) *ParseError {
	pe := &ParseError{
		BaseError: BaseError{
			Type:    ErrTypeParse,
			Message: err.Error(),
			Cause:   err,
		},
	}

	// Try to parse position from error message (format: "line:col: message")
	// Participle errors always include position in this format
	pe.Pos = parsePositionFromMessage(err.Error())

	return pe
}

// parsePositionFromMessage extracts position from error message format "line:col: message"
func parsePositionFromMessage(msg string) Position {
	var line, col int
	// Try parsing "line:col: message" format
	n, _ := fmt.Sscanf(msg, "%d:%d:", &line, &col)
	if n == 2 {
		return Position{Line: line, Column: col}
	}
	return Position{}
}

// ValidationErrorCode categorizes validation errors.
type ValidationErrorCode int

const (
	ErrCodeUnknownObject ValidationErrorCode = iota
	ErrCodeUnknownField
	ErrCodeUnknownLookup
	ErrCodeUnknownRelationship
	ErrCodeTypeMismatch
	ErrCodeInvalidOperator
	ErrCodeInvalidAggregation
	ErrCodeFieldNotFilterable
	ErrCodeFieldNotSortable
	ErrCodeFieldNotGroupable
	ErrCodeFieldNotAggregatable
	ErrCodeNestedSubqueryNotAllowed
	ErrCodeTooManyLookupLevels
	ErrCodeInvalidExpression
	ErrCodeMissingRequiredClause
	ErrCodeInvalidDateLiteral
	ErrCodeInvalidPagination
	ErrCodeWhereSubquerySingleField    // WHERE subquery must select exactly one field
	ErrCodeWhereSubqueryAggregateField // WHERE subquery cannot use aggregates
)

func (c ValidationErrorCode) String() string {
	switch c {
	case ErrCodeUnknownObject:
		return "UnknownObject"
	case ErrCodeUnknownField:
		return "UnknownField"
	case ErrCodeUnknownLookup:
		return "UnknownLookup"
	case ErrCodeUnknownRelationship:
		return "UnknownRelationship"
	case ErrCodeTypeMismatch:
		return "TypeMismatch"
	case ErrCodeInvalidOperator:
		return "InvalidOperator"
	case ErrCodeInvalidAggregation:
		return "InvalidAggregation"
	case ErrCodeFieldNotFilterable:
		return "FieldNotFilterable"
	case ErrCodeFieldNotSortable:
		return "FieldNotSortable"
	case ErrCodeFieldNotGroupable:
		return "FieldNotGroupable"
	case ErrCodeFieldNotAggregatable:
		return "FieldNotAggregatable"
	case ErrCodeNestedSubqueryNotAllowed:
		return "NestedSubqueryNotAllowed"
	case ErrCodeTooManyLookupLevels:
		return "TooManyLookupLevels"
	case ErrCodeInvalidExpression:
		return "InvalidExpression"
	case ErrCodeMissingRequiredClause:
		return "MissingRequiredClause"
	case ErrCodeInvalidDateLiteral:
		return "InvalidDateLiteral"
	case ErrCodeInvalidPagination:
		return "InvalidPagination"
	case ErrCodeWhereSubquerySingleField:
		return "WhereSubquerySingleField"
	case ErrCodeWhereSubqueryAggregateField:
		return "WhereSubqueryAggregateField"
	default:
		return "UnknownValidationError"
	}
}

// ValidationError represents a semantic error during validation.
type ValidationError struct {
	BaseError
	Code   ValidationErrorCode
	Object string // Object name (if applicable)
	Field  string // Field name (if applicable)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(code ValidationErrorCode, message string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: message,
		},
		Code: code,
	}
}

// NewValidationErrorWithPos creates a new ValidationError with position.
func NewValidationErrorWithPos(code ValidationErrorCode, pos Position, message string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: message,
			Pos:     pos,
		},
		Code: code,
	}
}

// UnknownObjectError creates an error for unknown object.
func UnknownObjectError(object string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown object: %s", object),
		},
		Code:   ErrCodeUnknownObject,
		Object: object,
	}
}

// UnknownFieldError creates an error for unknown field.
func UnknownFieldError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown field: %s.%s", object, field),
		},
		Code:   ErrCodeUnknownField,
		Object: object,
		Field:  field,
	}
}

// UnknownFieldErrorAt creates an error for unknown field with position.
func UnknownFieldErrorAt(object, field string, pos lexer.Position) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown field: %s.%s", object, field),
			Pos:     PosFromLexer(pos),
		},
		Code:   ErrCodeUnknownField,
		Object: object,
		Field:  field,
	}
}

// UnknownLookupError creates an error for unknown lookup relationship.
func UnknownLookupError(object, lookup string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown lookup: %s.%s", object, lookup),
		},
		Code:   ErrCodeUnknownLookup,
		Object: object,
		Field:  lookup,
	}
}

// UnknownLookupErrorAt creates an error for unknown lookup relationship with position.
func UnknownLookupErrorAt(object, lookup string, pos lexer.Position) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown lookup: %s.%s", object, lookup),
			Pos:     PosFromLexer(pos),
		},
		Code:   ErrCodeUnknownLookup,
		Object: object,
		Field:  lookup,
	}
}

// UnknownRelationshipError creates an error for unknown relationship.
func UnknownRelationshipError(object, relationship string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown relationship: %s.%s", object, relationship),
		},
		Code:   ErrCodeUnknownRelationship,
		Object: object,
		Field:  relationship,
	}
}

// UnknownRelationshipErrorAt creates an error for unknown relationship with position.
func UnknownRelationshipErrorAt(object, relationship string, pos lexer.Position) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("unknown relationship: %s.%s", object, relationship),
			Pos:     PosFromLexer(pos),
		},
		Code:   ErrCodeUnknownRelationship,
		Object: object,
		Field:  relationship,
	}
}

// TypeMismatchError creates an error for type mismatch.
func TypeMismatchError(expected, got FieldType) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("type mismatch: expected %s, got %s", expected, got),
		},
		Code: ErrCodeTypeMismatch,
	}
}

// FieldNotFilterableError creates an error for non-filterable field.
func FieldNotFilterableError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("field %s.%s is not filterable", object, field),
		},
		Code:   ErrCodeFieldNotFilterable,
		Object: object,
		Field:  field,
	}
}

// FieldNotSortableError creates an error for non-sortable field.
func FieldNotSortableError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("field %s.%s is not sortable", object, field),
		},
		Code:   ErrCodeFieldNotSortable,
		Object: object,
		Field:  field,
	}
}

// FieldNotGroupableError creates an error for non-groupable field.
func FieldNotGroupableError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("field %s.%s is not groupable", object, field),
		},
		Code:   ErrCodeFieldNotGroupable,
		Object: object,
		Field:  field,
	}
}

// WhereSubquerySingleFieldError creates an error when WHERE subquery selects more than one field.
func WhereSubquerySingleFieldError(pos lexer.Position) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: "WHERE subquery must select exactly one field",
			Pos:     PosFromLexer(pos),
		},
		Code: ErrCodeWhereSubquerySingleField,
	}
}

// WhereSubqueryAggregateFieldError creates an error when WHERE subquery uses aggregates.
func WhereSubqueryAggregateFieldError(pos lexer.Position) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: "WHERE subquery cannot use aggregate functions",
			Pos:     PosFromLexer(pos),
		},
		Code: ErrCodeWhereSubqueryAggregateField,
	}
}

// AccessError represents an access denied error.
type AccessError struct {
	BaseError
	Object string // Object name
	Field  string // Field name (empty if object-level error)
}

// NewAccessError creates an access error for an object.
func NewAccessError(object string) *AccessError {
	return &AccessError{
		BaseError: BaseError{
			Type:    ErrTypeAccess,
			Message: fmt.Sprintf("access denied to object: %s", object),
		},
		Object: object,
	}
}

// NewFieldAccessError creates an access error for a field.
func NewFieldAccessError(object, field string) *AccessError {
	return &AccessError{
		BaseError: BaseError{
			Type:    ErrTypeAccess,
			Message: fmt.Sprintf("access denied to field: %s.%s", object, field),
		},
		Object: object,
		Field:  field,
	}
}

// LimitType categorizes limit violations.
type LimitType int

const (
	LimitTypeMaxFields LimitType = iota
	LimitTypeMaxRecords
	LimitTypeMaxSubqueries
	LimitTypeMaxLookupDepth
	LimitTypeMaxSubqueryRecords
	LimitTypeMaxOffset
	LimitTypeMaxQueryLength
)

func (t LimitType) String() string {
	switch t {
	case LimitTypeMaxFields:
		return "MaxFields"
	case LimitTypeMaxRecords:
		return "MaxRecords"
	case LimitTypeMaxSubqueries:
		return "MaxSubqueries"
	case LimitTypeMaxLookupDepth:
		return "MaxLookupDepth"
	case LimitTypeMaxSubqueryRecords:
		return "MaxSubqueryRecords"
	case LimitTypeMaxOffset:
		return "MaxOffset"
	case LimitTypeMaxQueryLength:
		return "MaxQueryLength"
	default:
		return "UnknownLimit"
	}
}

// LimitError represents a limit exceeded error.
type LimitError struct {
	BaseError
	LimitType LimitType
	Limit     int // The configured limit
	Actual    int // The actual value that exceeded the limit
}

// NewLimitError creates a new LimitError.
func NewLimitError(limitType LimitType, limit, actual int) *LimitError {
	return &LimitError{
		BaseError: BaseError{
			Type:    ErrTypeLimit,
			Message: fmt.Sprintf("%s limit exceeded: %d (max: %d)", limitType, actual, limit),
		},
		LimitType: limitType,
		Limit:     limit,
		Actual:    actual,
	}
}

// ExecutionError represents a runtime error during query execution.
type ExecutionError struct {
	BaseError
	SQL      string // The SQL that caused the error (if available)
	SQLError error  // The underlying database error
}

// NewExecutionError creates a new ExecutionError.
func NewExecutionError(message string, sqlErr error) *ExecutionError {
	return &ExecutionError{
		BaseError: BaseError{
			Type:    ErrTypeExecution,
			Message: message,
			Cause:   sqlErr,
		},
		SQLError: sqlErr,
	}
}

// NewExecutionErrorWithSQL creates a new ExecutionError with SQL context.
func NewExecutionErrorWithSQL(message, sql string, sqlErr error) *ExecutionError {
	return &ExecutionError{
		BaseError: BaseError{
			Type:    ErrTypeExecution,
			Message: message,
			Cause:   sqlErr,
		},
		SQL:      sql,
		SQLError: sqlErr,
	}
}

// Error type checking helpers

// IsParseError checks if the error is a ParseError.
func IsParseError(err error) bool {
	var e *ParseError
	return errors.As(err, &e)
}

// IsValidationError checks if the error is a ValidationError.
func IsValidationError(err error) bool {
	var e *ValidationError
	return errors.As(err, &e)
}

// IsAccessError checks if the error is an AccessError.
func IsAccessError(err error) bool {
	var e *AccessError
	return errors.As(err, &e)
}

// IsLimitError checks if the error is a LimitError.
func IsLimitError(err error) bool {
	var e *LimitError
	return errors.As(err, &e)
}

// IsExecutionError checks if the error is an ExecutionError.
func IsExecutionError(err error) bool {
	var e *ExecutionError
	return errors.As(err, &e)
}
