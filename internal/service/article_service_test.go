package service

import (
	"context"
	"errors"
	"testing"
	"time"

	customErrors "github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/internal/mocks"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewArticleService(t *testing.T) {
	s := NewArticleService(nil, nil)
	require.NotNil(t, s)
}

func TestArticleService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	mockArticleRepo := mocks.NewMockArticleRepository(ctrl)

	articleService := &articleService{
		authorRepository:  mockAuthorRepo,
		articleRepository: mockArticleRepo,
	}

	t.Run("invalid author id format", func(t *testing.T) {
		req := &model.CreateArticleRequest{
			AuthorID: "not-a-uuid",
			Title:    "Some Title",
			Body:     "Some Body",
		}

		res, err := articleService.Create(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), customErrors.ErrInvalidData.Error())
		assert.Nil(t, res)
	})

	t.Run("author not found", func(t *testing.T) {
		authorID := uuid.New()

		mockAuthorRepo.EXPECT().
			FindByID(gomock.Any(), authorID).
			Return(nil, nil)

		req := &model.CreateArticleRequest{
			AuthorID: authorID.String(),
			Title:    "Some Title",
			Body:     "Some Body",
		}

		res, err := articleService.Create(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), customErrors.ErrRecordNotFound.Error())
		assert.Nil(t, res)
	})

	t.Run("error from author repo", func(t *testing.T) {
		authorID := uuid.New()

		mockAuthorRepo.EXPECT().
			FindByID(gomock.Any(), authorID).
			Return(nil, errors.New("db error"))

		req := &model.CreateArticleRequest{
			AuthorID: authorID.String(),
			Title:    "Some Title",
			Body:     "Some Body",
		}

		res, err := articleService.Create(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
		assert.Nil(t, res)
	})

	t.Run("article repo insert error", func(t *testing.T) {
		authorID := uuid.New()

		mockAuthorRepo.EXPECT().
			FindByID(gomock.Any(), authorID).
			Return(&model.Author{ID: authorID, Name: "Test"}, nil)

		mockArticleRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("insert error"))

		req := &model.CreateArticleRequest{
			AuthorID: authorID.String(),
			Title:    "Some Title",
			Body:     "Some Body",
		}

		res, err := articleService.Create(ctx, req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert error")
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		authorID := uuid.New()

		mockAuthorRepo.EXPECT().
			FindByID(gomock.Any(), authorID).
			Return(&model.Author{ID: authorID, Name: "Test"}, nil)

		mockArticleRepo.EXPECT().
			Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, a *model.Article) (*model.Article, error) {
				a.CreatedAt = time.Now()
				return a, nil
			})

		req := &model.CreateArticleRequest{
			AuthorID: authorID.String(),
			Title:    "Some Title",
			Body:     "Some Body",
		}

		res, err := articleService.Create(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, req.Title, res.Title)
		assert.Equal(t, req.Body, res.Body)
		assert.Equal(t, authorID, res.AuthorID)
	})
}

func TestArticleService_FindAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockAuthorRepo := mocks.NewMockAuthorRepository(ctrl)
	mockArticleRepo := mocks.NewMockArticleRepository(ctrl)

	articleService := &articleService{
		authorRepository:  mockAuthorRepo,
		articleRepository: mockArticleRepo,
	}

	t.Run("success", func(t *testing.T) {
		expected := []*model.Article{
			{
				ID:        uuid.New(),
				AuthorID:  uuid.New(),
				Author:    "Author Name",
				Title:     "Article Title",
				Body:      "Article Body",
				CreatedAt: time.Now(),
			},
		}

		mockArticleRepo.EXPECT().
			FindAll(gomock.Any(), gomock.Any()).
			Return(expected, 1, nil)

		res, total, err := articleService.FindAll(ctx, model.ArticleQuery{
			Query: "keyword",
			Page:  1,
			Limit: 10,
		})

		assert.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Equal(t, expected, res)
	})

	t.Run("empty articles", func(t *testing.T) {
		expected := []*model.Article{}
		mockArticleRepo.EXPECT().
			FindAll(gomock.Any(), gomock.Any()).
			Return(nil, 0, nil)

		res, total, err := articleService.FindAll(ctx, model.ArticleQuery{
			Query: "keyword",
			Page:  1,
			Limit: 10,
		})

		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		assert.Equal(t, 0, total)
	})

	t.Run("error from repo", func(t *testing.T) {
		mockArticleRepo.EXPECT().
			FindAll(gomock.Any(), gomock.Any()).
			Return(nil, 0, errors.New("repo error"))

		res, total, err := articleService.FindAll(ctx, model.ArticleQuery{
			Query: "keyword",
			Page:  1,
			Limit: 10,
		})

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, 0, total)
	})
}
