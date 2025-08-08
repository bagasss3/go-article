package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	customErr "github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockAuthorService struct {
	mock.Mock
}

func (m *MockAuthorService) Create(ctx context.Context, req *model.CreateAuthorRequest) (*model.Author, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.Author), args.Error(1)
}

func (m *MockAuthorService) FindByID(ctx context.Context, id string) (*model.Author, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Author), args.Error(1)
}

func TestAuthorHandler_Create(t *testing.T) {
	e := echo.New()
	e.Validator = &model.CustomValidator{Validator: validator.New()}

	t.Run("success", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		body := `{"name":"Jane Doe"}`
		req := httptest.NewRequest(http.MethodPost, "/author", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expected := &model.Author{
			ID:   uuid.New(),
			Name: "Jane Doe",
		}
		service.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		body := `{"name":"invalid-json"`
		req := httptest.NewRequest(http.MethodPost, "/author", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		body := `{"name":""}`
		req := httptest.NewRequest(http.MethodPost, "/author", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		body := `{"name":"Jane"}`
		req := httptest.NewRequest(http.MethodPost, "/author", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var dummy *model.Author
		service.On("Create", mock.Anything, mock.Anything).Return(dummy, echo.NewHTTPError(http.StatusInternalServerError, "fail"))

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestAuthorHandler_GetByID(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		authorID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/author/"+authorID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(authorID)

		expected := &model.Author{
			ID:   uuid.MustParse(authorID),
			Name: "Jane Doe",
		}
		service.On("FindByID", mock.Anything, authorID).Return(expected, nil)

		err := handler.getByID(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		service := new(MockAuthorService)
		handler := NewAuthorHandler(service)

		authorID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, "/author/"+authorID, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(authorID)

		var dummy *model.Author
		service.On("FindByID", mock.Anything, authorID).Return(dummy, echo.NewHTTPError(http.StatusInternalServerError, "fail"))

		err := handler.getByID(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestAuthorHandler_Register(t *testing.T) {
	service := new(MockAuthorService)
	handler := NewAuthorHandler(service)

	e := echo.New()
	g := e.Group("/api")

	handler.Register(g)

	routes := e.Routes()
	require.True(t, len(routes) > 0)

	// Check for specific routes
	foundPostRoute := false
	foundGetRoute := false
	for _, route := range routes {
		if route.Method == "POST" && route.Path == "/api/author" {
			foundPostRoute = true
		}
		if route.Method == "GET" && route.Path == "/api/author/:id" {
			foundGetRoute = true
		}
	}
	require.True(t, foundPostRoute, "POST route should be registered")
	require.True(t, foundGetRoute, "GET route should be registered")
}

func TestHandleError(t *testing.T) {
	e := echo.New()

	t.Run("custom error with existing status code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		custErr := &customErr.CustomError{
			Message:          customErr.ErrInvalidData,
			MessageDeveloper: "Developer message for bad request",
		}

		err := handleError(c, custErr)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("custom error with non-existing status code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		unknownErr := errors.New("unknown custom error")
		custErr := &customErr.CustomError{
			Message:          unknownErr,
			MessageDeveloper: "Developer message for unknown error",
		}

		err := handleError(c, custErr)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("non-custom error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		regularErr := errors.New("regular error message")
		err := handleError(c, regularErr)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("echo http error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		httpErr := echo.NewHTTPError(http.StatusNotFound, "not found")
		err := handleError(c, httpErr)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
