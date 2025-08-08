package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArticleRepository_Create(t *testing.T) {
	kit := initializeRepoTestKit(t)
	defer kit.closer()

	repo := NewArticleRepository(kit.db, kit.cache)

	ctx := context.TODO()
	article := &model.Article{
		AuthorID: uuid.New(),
		Title:    "Test Title",
		Body:     "Test Body",
	}

	t.Run("success", func(t *testing.T) {
		kit.mock.ExpectQuery("INSERT INTO articles").
			WithArgs(sqlmock.AnyArg(), article.AuthorID, article.Title, article.Body).
			WillReturnRows(sqlmock.NewRows([]string{"created_at"}).AddRow(time.Now()))

		result, err := repo.Create(ctx, article)
		require.NoError(t, err)
		require.NotNil(t, result)

		key := "article:page=1:limit=10"
		err = kit.cache.Get(ctx, key, &struct {
			Results []*model.Article
			Total   int
		}{})
		assert.Error(t, err)
		assert.EqualError(t, err, "cache miss")
	})

	t.Run("insert error", func(t *testing.T) {
		kit.mock.ExpectQuery("INSERT INTO articles").
			WithArgs(sqlmock.AnyArg(), article.AuthorID, article.Title, article.Body).
			WillReturnError(errors.New("db error"))

		_, err := repo.Create(ctx, article)
		require.Error(t, err)
	})
}

func TestArticleRepository_FindAll(t *testing.T) {
	kit := initializeRepoTestKit(t)
	defer kit.closer()

	repo := NewArticleRepository(kit.db, kit.cache)
	ctx := context.TODO()

	t.Run("success with result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "author_id", "name", "title", "body", "created_at"}).
			AddRow(uuid.New(), uuid.New(), "Author", "Title", "Body", time.Now())

		kit.mock.ExpectQuery("SELECT a.id, a.author_id").
			WillReturnRows(rows)

		kit.mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		filter := model.ArticleQuery{Page: 1, Limit: 10}
		res, total, err := repo.FindAll(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})

	t.Run("query error", func(t *testing.T) {
		kit.mock.ExpectQuery("SELECT a.id, a.author_id").
			WillReturnError(errors.New("query error"))

		_, _, err := repo.FindAll(ctx, model.ArticleQuery{})
		require.Error(t, err)
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "author_id", "name", "title", "body", "created_at"}).
			AddRow("bad-uuid", uuid.New(), "Author", "Title", "Body", time.Now())

		kit.mock.ExpectQuery("SELECT a.id, a.author_id").
			WillReturnRows(rows)

		_, _, err := repo.FindAll(ctx, model.ArticleQuery{})
		require.Error(t, err)
	})

	t.Run("count error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "author_id", "name", "title", "body", "created_at"}).
			AddRow(uuid.New(), uuid.New(), "Author", "Title", "Body", time.Now())

		kit.mock.ExpectQuery("SELECT a.id, a.author_id").
			WillReturnRows(rows)

		kit.mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.FindAll(ctx, model.ArticleQuery{})
		require.Error(t, err)
	})

	t.Run("with title/body filter and author name", func(t *testing.T) {
		titleBody := "%search%"
		authorName := "%john%"

		filter := model.ArticleQuery{
			Query:  "search",
			Author: "john",
			Page:   2,
			Limit:  5,
		}

		rows := sqlmock.NewRows([]string{"id", "author_id", "name", "title", "body", "created_at"}).
			AddRow(uuid.New(), uuid.New(), "John", "Search match", "Body", time.Now())

		kit.mock.ExpectQuery("SELECT a.id, a.author_id.*FROM articles a.*JOIN authors au").
			WithArgs(titleBody, titleBody, authorName, 5, 5).
			WillReturnRows(rows)

		kit.mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM articles a.*").
			WithArgs(titleBody, titleBody, authorName).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		res, total, err := repo.FindAll(ctx, filter)
		require.NoError(t, err)
		require.Len(t, res, 1)
		assert.Equal(t, 1, total)
	})

	t.Run("default limit and page", func(t *testing.T) {
		filter := model.ArticleQuery{}

		rows := sqlmock.NewRows([]string{"id", "author_id", "name", "title", "body", "created_at"}).
			AddRow(uuid.New(), uuid.New(), "John", "Title", "Body", time.Now())

		kit.mock.ExpectQuery("SELECT a.id, a.author_id").
			WithArgs(10, 0).
			WillReturnRows(rows)

		kit.mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		res, total, err := repo.FindAll(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, res, 1)
	})
}
