package engine

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

// ErrorType categorizes DML errors.
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

// Position represents a location in the DML statement string.
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

// BaseError is the base error type for all DML errors.
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
	pe.Pos = parsePositionFromMessage(err.Error())

	return pe
}

// parsePositionFromMessage extracts position from error message format "line:col: message"
func parsePositionFromMessage(msg string) Position {
	var line, col int
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
	ErrCodeTypeMismatch
	ErrCodeReadOnlyField
	ErrCodeMissingRequired
	ErrCodeInvalidExpression
	ErrCodeInvalidValue
	ErrCodeExternalIdNotFound
	ErrCodeExternalIdNotUnique
	ErrCodeDeleteRequiresWhere
	ErrCodeDuplicateField
)

func (c ValidationErrorCode) String() string {
	switch c {
	case ErrCodeUnknownObject:
		return "UnknownObject"
	case ErrCodeUnknownField:
		return "UnknownField"
	case ErrCodeTypeMismatch:
		return "TypeMismatch"
	case ErrCodeReadOnlyField:
		return "ReadOnlyField"
	case ErrCodeMissingRequired:
		return "MissingRequired"
	case ErrCodeInvalidExpression:
		return "InvalidExpression"
	case ErrCodeInvalidValue:
		return "InvalidValue"
	case ErrCodeExternalIdNotFound:
		return "ExternalIdNotFound"
	case ErrCodeExternalIdNotUnique:
		return "ExternalIdNotUnique"
	case ErrCodeDeleteRequiresWhere:
		return "DeleteRequiresWhere"
	case ErrCodeDuplicateField:
		return "DuplicateField"
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

// ReadOnlyFieldError creates an error for attempting to write a read-only field.
func ReadOnlyFieldError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("field %s.%s is read-only", object, field),
		},
		Code:   ErrCodeReadOnlyField,
		Object: object,
		Field:  field,
	}
}

// MissingRequiredFieldError creates an error for missing required field.
func MissingRequiredFieldError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("required field %s.%s is missing", object, field),
		},
		Code:   ErrCodeMissingRequired,
		Object: object,
		Field:  field,
	}
}

// TypeMismatchError creates an error for type mismatch.
func TypeMismatchError(object, field string, expected, got FieldType) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("type mismatch for %s.%s: expected %s, got %s", object, field, expected, got),
		},
		Code:   ErrCodeTypeMismatch,
		Object: object,
		Field:  field,
	}
}

// ExternalIdNotFoundError creates an error when external ID field doesn't exist.
func ExternalIdNotFoundError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("external ID field %s.%s not found or not marked as external ID", object, field),
		},
		Code:   ErrCodeExternalIdNotFound,
		Object: object,
		Field:  field,
	}
}

// DeleteRequiresWhereError creates an error when DELETE has no WHERE clause.
func DeleteRequiresWhereError(object string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("DELETE from %s requires a WHERE clause for safety", object),
		},
		Code:   ErrCodeDeleteRequiresWhere,
		Object: object,
	}
}

// DuplicateFieldError creates an error for duplicate field in field list.
func DuplicateFieldError(object, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			Type:    ErrTypeValidation,
			Message: fmt.Sprintf("duplicate field %s.%s in field list", object, field),
		},
		Code:   ErrCodeDuplicateField,
		Object: object,
		Field:  field,
	}
}

// AccessError represents an access denied error.
type AccessError struct {
	BaseError
	Object    string    // Object name
	Field     string    // Field name (empty if object-level error)
	Operation Operation // The operation that was denied
}

// NewWriteAccessError creates an access error for write operation on an object.
func NewWriteAccessError(object string, op Operation) *AccessError {
	return &AccessError{
		BaseError: BaseError{
			Type:    ErrTypeAccess,
			Message: fmt.Sprintf("%s access denied to object: %s", op, object),
		},
		Object:    object,
		Operation: op,
	}
}

// NewFieldWriteAccessError creates an access error for write operation on a field.
func NewFieldWriteAccessError(object, field string) *AccessError {
	return &AccessError{
		BaseError: BaseError{
			Type:    ErrTypeAccess,
			Message: fmt.Sprintf("write access denied to field: %s.%s", object, field),
		},
		Object: object,
		Field:  field,
	}
}

// LimitType categorizes limit violations.
type LimitType int

const (
	LimitTypeMaxBatchSize LimitType = iota
	LimitTypeMaxFieldsPerRow
	LimitTypeMaxStatementLength
)

func (t LimitType) String() string {
	switch t {
	case LimitTypeMaxBatchSize:
		return "MaxBatchSize"
	case LimitTypeMaxFieldsPerRow:
		return "MaxFieldsPerRow"
	case LimitTypeMaxStatementLength:
		return "MaxStatementLength"
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

// ExecutionError represents a runtime error during DML execution.
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

// RuleValidationError represents one or more validation rule failures.
type RuleValidationError struct {
	Rules []ValidationRuleError
}

func (e *RuleValidationError) Error() string {
	if len(e.Rules) == 0 {
		return "validation rule failed"
	}
	if len(e.Rules) == 1 {
		return fmt.Sprintf("validation rule failed: %s", e.Rules[0].Message)
	}
	return fmt.Sprintf("validation rules failed: %d violations", len(e.Rules))
}

// IsRuleValidationError checks if the error is a RuleValidationError.
func IsRuleValidationError(err error) bool {
	var e *RuleValidationError
	return errors.As(err, &e)
}

// DefaultEvalError represents a failure evaluating a default expression.
type DefaultEvalError struct {
	Field      string
	Expression string
	Cause      error
}

func (e *DefaultEvalError) Error() string {
	return fmt.Sprintf("failed to evaluate default for field %s (expression: %s): %s",
		e.Field, e.Expression, e.Cause)
}

func (e *DefaultEvalError) Unwrap() error {
	return e.Cause
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
