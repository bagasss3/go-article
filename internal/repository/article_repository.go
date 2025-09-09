package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type articleRepository struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewArticleRepository(db *gorm.DB, cache cache.Cache) model.ArticleRepository {
	return &articleRepository{
		db:    db,
		cache: cache,
	}
}

func (r *articleRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Article, error) {
	articleKey := fmt.Sprintf("%s:detail:%s", model.ArticleKey, id.String())
	var article model.Article
	if err := r.cache.Get(ctx, articleKey, &article); err == nil {
		return &article, nil
	}

	err := r.db.WithContext(ctx).Take(&article, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		log.Error(err)
		return nil, err
	}

	if err := r.cache.Set(ctx, articleKey, &article, config.RedisExpired()); err != nil {
		log.Warn("failed to cache individual article")
	}

	return &article, nil
}

func (r *articleRepository) FindAll(ctx context.Context, filter model.ArticleQuery) ([]*model.Article, int, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (filter.Page - 1) * limit
	if filter.Page <= 0 {
		offset = 0
	}

	cacheKey := fmt.Sprintf("%s:%d:%d", model.ArticleKey, filter.Page, limit)
	var cachedIDs []uuid.UUID
	if filter.Query == "" && filter.Author == "" {
		if err := r.cache.Get(ctx, cacheKey, &cachedIDs); err == nil {
			var wg sync.WaitGroup
			var mu sync.Mutex
			var results []*model.Article

			for _, id := range cachedIDs {
				wg.Add(1)
				go func(articleID uuid.UUID) {
					defer wg.Done()
					article, err := r.FindByID(ctx, articleID)
					if err == nil && article != nil {
						mu.Lock()
						results = append(results, article)
						mu.Unlock()
					}
				}(id)
			}
			wg.Wait()

			if len(results) > 0 {
				totalKey := fmt.Sprintf("%s:total", model.ArticleKey)
				var total int64
				if err := r.cache.Get(ctx, totalKey, &total); err == nil {
					return results, int(total), nil
				}
			}
		}
	}

	query := r.db.WithContext(ctx).Model(&model.Article{})

	if filter.Query != "" {
		query = query.Where("to_tsvector('simple', title || ' ' || body) @@ plainto_tsquery('simple', ?)", filter.Query)
	}

	if filter.Author != "" {
		var authorIDs []uuid.UUID
		err := r.db.WithContext(ctx).Model(&model.Author{}).
			Where("to_tsvector('simple', name) @@ plainto_tsquery('simple', ?)", filter.Author).
			Pluck("id", &authorIDs).Error
		if err != nil {
			log.Error(err)
			return nil, 0, err
		}
		if len(authorIDs) == 0 {
			return []*model.Article{}, 0, nil
		}
		query = query.Where("author_id IN ?", authorIDs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Error(err)
		return nil, 0, err
	}

	var results []*model.Article
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&results).Error
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}

	if filter.Query == "" && filter.Author == "" {
		for _, article := range results {
			articleKey := fmt.Sprintf("%s:detail:%s", model.ArticleKey, article.ID.String())
			if err := r.cache.Set(ctx, articleKey, article, config.RedisExpired()); err != nil {
				log.Warn("failed to cache individual article")
			}
		}

		var articleIDs []uuid.UUID
		for _, article := range results {
			articleIDs = append(articleIDs, article.ID)
		}
		if err := r.cache.Set(ctx, cacheKey, articleIDs, config.RedisExpired()); err != nil {
			log.Warn("failed to cache article IDs")
		}

		totalKey := fmt.Sprintf("%s:total", model.ArticleKey)
		if err := r.cache.Set(ctx, totalKey, total, config.RedisExpired()); err != nil {
			log.Warn("failed to cache total count")
		}
	}

	return results, int(total), nil
}

func (r *articleRepository) Create(ctx context.Context, article *model.Article) (*model.Article, error) {
	article.ID = uuid.New()

	result := r.db.WithContext(ctx).Table("articles").Create(map[string]interface{}{
		"id":         article.ID,
		"author_id":  article.AuthorID,
		"title":      article.Title,
		"body":       article.Body,
		"created_at": "NOW()",
	})

	if result.Error != nil {
		log.Error(result.Error)
		return nil, result.Error
	}

	var createdArticle model.Article
	err := r.db.WithContext(ctx).Table("articles").
		Where("id = ?", article.ID).
		Select("created_at").
		Scan(&createdArticle).Error
	if err != nil {
		log.Error(err)
		return nil, err
	}
	article.CreatedAt = createdArticle.CreatedAt

	totalKey := fmt.Sprintf("%s:total", model.ArticleKey)
	if err := r.cache.Delete(ctx, totalKey); err != nil {
		log.Warn("failed to delete total count cache")
	}

	for page := 1; page <= 5; page++ {
		for limit := 10; limit <= 50; limit += 10 {
			cacheKey := fmt.Sprintf("%s:%d:%d", model.ArticleKey, page, limit)
			if err := r.cache.Delete(ctx, cacheKey); err != nil {
				log.Warn("failed to delete pagination cache")
			}
		}
	}

	return article, nil
}
