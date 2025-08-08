package handler

import (
	"net/http"

	customErr "github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/pkg/response"
	"github.com/labstack/echo/v4"
)

func handleError(c echo.Context, err error) error {
	// Check if it's a `CustomError`
	if custErr, ok := err.(*customErr.CustomError); ok {
		// Get the corresponding HTTP status code
		statusCode, exists := customErr.ErrorStatusMap[custErr.Message]
		if !exists {
			statusCode = http.StatusInternalServerError
		}

		// Send structured error response
		return response.ResponseInterfaceError(c, statusCode, custErr.MessageDeveloper, custErr.Message.Error())
	}

	// Default fallback for unknown errors
	return response.ResponseInterfaceError(c, http.StatusInternalServerError, err.Error(), customErr.ErrInternalServer.Error())
}
