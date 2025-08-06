package server

import (
	"fmt"
	"net/http"

	"github.com/bagasss3/go-article/pkg/model"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type HTTPServer struct {
	echo *echo.Echo
}

func NewHTTPServer() *HTTPServer {
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	validator := validator.New()
	e.Validator = &model.CustomValidator{Validator: validator}
	e.HTTPErrorHandler = customHTTPErrorHandler

	return &HTTPServer{echo: e}
}

func (s *HTTPServer) Engine() *echo.Echo {
	return s.echo
}

func (s *HTTPServer) Start(port string) error {
	return s.echo.Start(":" + port)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	// Generate a unique request ID (use a library or header for real-world scenarios)
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)
	if requestID == "" {
		requestID = "unknown"
	}

	// Default error response
	response := model.JsonResponsError{
		RequestId:    requestID,
		StatusCode:   http.StatusInternalServerError,
		ErrorCode:    http.StatusInternalServerError,
		ErrorMessage: "Internal Server Error",
	}

	// Customize based on error type
	if he, ok := err.(*echo.HTTPError); ok {
		response.StatusCode = he.Code
		if he.Message != nil {
			response.ErrorMessage = fmt.Sprintf("%v", he.Message)
		}
		if he.Internal != nil {
			response.DeveloperMessage = he.Internal.Error()
		}
	} else {
		// For non-HTTP errors, use the default response
		response.DeveloperMessage = err.Error()
	}

	// Send the JSON response
	if !c.Response().Committed {
		c.JSON(response.StatusCode, response)
	}
}
