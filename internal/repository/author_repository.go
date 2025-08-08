package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
)

type authorRepository struct {
	db    *sql.DB
	cache cache.Cache
}

func NewAuthorRepository(db *sql.DB, cache cache.Cache) model.AuthorRepository {
	return &authorRepository{
		db:    db,
		cache: cache,
	}
}

func (r *authorRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	var author model.Author

	key := fmt.Sprintf("%s:%s", model.AuthorKey, id.String())
	if err := r.cache.Get(ctx, key, &author); err == nil {
		return &author, nil
	}

	query := `SELECT id, name FROM authors WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&author.ID, &author.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	err = r.cache.Set(ctx, key, &author, config.RedisExpired())
	if err != nil {

	}

	return &author, nil
}

func (r *authorRepository) Create(ctx context.Context, author *model.Author) (*model.Author, error) {
	author.ID = uuid.New()

	query := `INSERT INTO authors (id, name) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, author.ID, author.Name)
	if err != nil {
		return nil, err
	}

	return author, nil
}
