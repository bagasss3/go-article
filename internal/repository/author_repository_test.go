package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestAuthorRepository_FindByID(t *testing.T) {
	kit := initializeRepoTestKit(t)
	defer kit.closer()

	repo := NewAuthorRepository(kit.db, kit.cache)
	ctx := context.TODO()
	authorID := uuid.New()
	cacheKey := model.AuthorKey + ":" + authorID.String()

	t.Run("found from db and cached", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(authorID, "John Doe")

		kit.mock.ExpectQuery("SELECT id, name FROM authors WHERE id =").
			WithArgs(authorID).
			WillReturnRows(rows)

		res, err := repo.FindByID(ctx, authorID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "John Doe", res.Name)

		var cached model.Author
		err = kit.cache.Get(ctx, cacheKey, &cached)
		require.NoError(t, err)
		require.Equal(t, "John Doe", cached.Name)
	})

	t.Run("found from cache", func(t *testing.T) {
		res, err := repo.FindByID(ctx, authorID)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "John Doe", res.Name)

		require.NoError(t, kit.mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		missingID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "name"}) 

		kit.mock.ExpectQuery("SELECT id, name FROM authors WHERE id =").
			WithArgs(missingID).
			WillReturnRows(rows)

		res, err := repo.FindByID(ctx, missingID)
		require.NoError(t, err)
		require.Nil(t, res)
	})

	t.Run("query error", func(t *testing.T) {
		brokenID := uuid.New()

		kit.mock.ExpectQuery("SELECT id, name FROM authors WHERE id =").
			WithArgs(brokenID).
			WillReturnError(errors.New("db error"))

		res, err := repo.FindByID(ctx, brokenID)
		require.Error(t, err)
		require.Nil(t, res)
	})
}

func TestAuthorRepository_Create(t *testing.T) {
	kit := initializeRepoTestKit(t)
	defer kit.closer()

	repo := NewAuthorRepository(kit.db, kit.cache)
	ctx := context.TODO()
	author := &model.Author{
		Name: "Jane Doe",
	}

	t.Run("success", func(t *testing.T) {
		kit.mock.ExpectExec("INSERT INTO authors").
			WithArgs(sqlmock.AnyArg(), author.Name).
			WillReturnResult(sqlmock.NewResult(1, 1))

		res, err := repo.Create(ctx, author)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, author.Name, res.Name)
		require.NotEqual(t, uuid.Nil, res.ID)
	})

	t.Run("insert error", func(t *testing.T) {
		kit.mock.ExpectExec("INSERT INTO authors").
			WithArgs(sqlmock.AnyArg(), author.Name).
			WillReturnError(errors.New("insert failed"))

		res, err := repo.Create(ctx, author)
		require.Error(t, err)
		require.Nil(t, res)
	})
}
