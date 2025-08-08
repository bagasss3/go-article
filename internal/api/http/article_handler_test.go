package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bagasss3/go-article/pkg/model"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockArticleService struct {
	mock.Mock
}

func (m *MockArticleService) FindAll(ctx context.Context, query model.ArticleQuery) ([]*model.Article, int, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*model.Article), args.Int(1), args.Error(2)
}

func (m *MockArticleService) Create(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*model.Article), args.Error(1)
}

func TestArticleHandler_Create(t *testing.T) {
	e := echo.New()
	validator := validator.New()
	e.Validator = &model.CustomValidator{Validator: validator}

	t.Run("success", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		authorID := uuid.New().String()
		body := `{"author_id":"` + authorID + `","title":"Test Title","body":"Test Body"}`
		req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expected := &model.Article{
			ID:        uuid.New(),
			AuthorID:  uuid.MustParse(authorID),
			Title:     "Test Title",
			Body:      "Test Body",
			CreatedAt: time.Now(),
		}
		service.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		body := `{"author_id":"invalid-json"`
		req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		body := `{"title":"T","body":"B"}`
		req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		authorID := uuid.New().String()
		body := `{"author_id":"` + authorID + `","title":"Test","body":"Content"}`
		req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		var dummy *model.Article
		service.On("Create", mock.Anything, mock.Anything).Return(dummy, echo.NewHTTPError(http.StatusInternalServerError, "fail"))
		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestArticleHandler_GetAll(t *testing.T) {
	e := echo.New()

	t.Run("success", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/article?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		expected := []*model.Article{
			{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				Author:    "John",
				Title:     "Test Title",
				Body:      "Content",
				CreatedAt: time.Now(),
			},
		}
		query := model.ArticleQuery{Page: 1, Limit: 10}
		service.On("FindAll", mock.Anything, query).Return(expected, 1, nil)

		err := handler.getAll(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("validation error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		body := `{"title":"T","body":"B"}`
		req := httptest.NewRequest(http.MethodPost, "/article", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.create(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("bind error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/article?page=bad", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler.getAll(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("service error", func(t *testing.T) {
		service := new(MockArticleService)
		handler := NewArticleHandler(service)

		req := httptest.NewRequest(http.MethodGet, "/article?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		query := model.ArticleQuery{Page: 1, Limit: 10}
		var dummy []*model.Article
		service.On("FindAll", mock.Anything, query).Return(dummy, 0, echo.NewHTTPError(http.StatusInternalServerError, "fail"))

		err := handler.getAll(c)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestArticleHandler_Register(t *testing.T) {
	service := new(MockArticleService)
	handler := NewArticleHandler(service)

	e := echo.New()
	g := e.Group("/api")

	handler.Register(g)

	routes := e.Routes()
	require.True(t, len(routes) > 0)

	foundGetRoute := false
	foundPostRoute := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/article" {
			foundGetRoute = true
		}
		if route.Method == "POST" && route.Path == "/api/article" {
			foundPostRoute = true
		}
	}
	require.True(t, foundGetRoute, "GET route should be registered")
	require.True(t, foundPostRoute, "POST route should be registered")
}
