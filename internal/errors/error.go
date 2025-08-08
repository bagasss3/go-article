package errors

import (
	"errors"
	"net/http"
)

type CustomError struct {
	Message          error  // The main error (like ErrRecordNotFound)
	MessageDeveloper string // Additional message
}

// Error method to satisfy the error interface
func (e *CustomError) Error() string {
	return e.Message.Error()
}

// Unwrap allows for error unwrapping
func (e *CustomError) Unwrap() error {
	return e.Message
}

// Common errors
var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrRecordNotFound   = errors.New("record not found")
	ErrDuplicate        = errors.New("record already exists")
	ErrInvalidData      = errors.New("invalid request")
	ErrInternalServer   = errors.New("internal server error")
)

// Default error messages for known errors
var errorMessages = map[error]string{
	ErrPermissionDenied: "permission denied",
	ErrUnauthorized:     "unauthorized access",
	ErrRecordNotFound:   "record not found",
	ErrDuplicate:        "record already exists",
	ErrInvalidData:      "invalid request data",
	ErrInternalServer:   "internal server error",
}

// New creates a `CustomError` with an optional dynamic developer message.
func New(errorType error, developerMessage string) *CustomError {
	if instance, exists := ErrorInstanceMap[errorType.Error()]; exists {
		errorType = instance
	}

	return &CustomError{
		Message:          errorType,
		MessageDeveloper: developerMessage,
	}
}

// GetDefaultMessage returns a default error message based on the type
func GetDefaultMessage(err error) string {
	if msg, exists := errorMessages[err]; exists {
		return msg
	}
	return "unexpected error occurred"
}

// Mapping HTTP Status Codes to Error Types
var ErrorStatusMap = map[error]int{
	ErrPermissionDenied: http.StatusForbidden,
	ErrUnauthorized:     http.StatusUnauthorized,
	ErrRecordNotFound:   http.StatusNotFound,
	ErrDuplicate:        http.StatusBadRequest,
	ErrInvalidData:      http.StatusBadRequest,
	ErrInternalServer:   http.StatusInternalServerError,
}

var ErrorInstanceMap = map[string]error{
	ErrPermissionDenied.Error(): ErrPermissionDenied,
	ErrUnauthorized.Error():     ErrUnauthorized,
	ErrRecordNotFound.Error():   ErrRecordNotFound,
	ErrDuplicate.Error():        ErrDuplicate,
	ErrInvalidData.Error():      ErrInvalidData,
	ErrInternalServer.Error():   ErrInternalServer,
}

func GetErrorByStatusCode(statusCode int) error {
	errorMapping := map[int]error{
		403: ErrPermissionDenied,
		401: ErrUnauthorized,
		404: ErrRecordNotFound,
		400: ErrInvalidData,
		409: ErrDuplicate,
		500: ErrInternalServer,
	}

	if err, exists := errorMapping[statusCode]; exists {
		return err
	}
	return ErrInternalServer // Default to internal server error
}
