package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
