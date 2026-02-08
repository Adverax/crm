package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Code string

const (
	CodeNotFound     Code = "NOT_FOUND"
	CodeBadRequest   Code = "BAD_REQUEST"
	CodeForbidden    Code = "FORBIDDEN"
	CodeUnauthorized Code = "UNAUTHORIZED"
	CodeConflict     Code = "CONFLICT"
	CodeInternal     Code = "INTERNAL"
	CodeValidation   Code = "VALIDATION"
)

type AppError struct {
	Code       Code   `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NotFound(entity string, id string) *AppError {
	return &AppError{
		Code:       CodeNotFound,
		Message:    fmt.Sprintf("%s with id %s not found", entity, id),
		HTTPStatus: http.StatusNotFound,
	}
}

func BadRequest(message string) *AppError {
	return &AppError{
		Code:       CodeBadRequest,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

func Forbidden(reason string) *AppError {
	return &AppError{
		Code:       CodeForbidden,
		Message:    reason,
		HTTPStatus: http.StatusForbidden,
	}
}

func Unauthorized(reason string) *AppError {
	return &AppError{
		Code:       CodeUnauthorized,
		Message:    reason,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func Conflict(message string) *AppError {
	return &AppError{
		Code:       CodeConflict,
		Message:    message,
		HTTPStatus: http.StatusConflict,
	}
}

func Internal(message string) *AppError {
	return &AppError{
		Code:       CodeInternal,
		Message:    message,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func Validation(message string) *AppError {
	return &AppError{
		Code:       CodeValidation,
		Message:    message,
		HTTPStatus: http.StatusBadRequest,
	}
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func Respond(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, errorResponse{
			Error: errorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	c.JSON(http.StatusInternalServerError, errorResponse{
		Error: errorBody{
			Code:    CodeInternal,
			Message: "internal server error",
		},
	})
}
