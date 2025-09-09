package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type authorRepository struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewAuthorRepository(db *gorm.DB, cache cache.Cache) model.AuthorRepository {
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

	err := r.db.WithContext(ctx).Table("authors").
		Where("id = ?", id).
		First(&author).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Error(err)
		return nil, err
	}

	if err := r.cache.Set(ctx, key, &author, config.RedisExpired()); err != nil {
		log.Warn("failed to cache author")
	}

	return &author, nil
}

func (r *authorRepository) Create(ctx context.Context, author *model.Author) (*model.Author, error) {
	author.ID = uuid.New()

	err := r.db.WithContext(ctx).Table("authors").Create(author).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return author, nil
}
