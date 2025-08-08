package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
)

type articleRepository struct {
	db    *sql.DB
	cache cache.Cache
}

func NewArticleRepository(db *sql.DB, cache cache.Cache) model.ArticleRepository {
	return &articleRepository{
		db:    db,
		cache: cache,
	}
}

func (r *articleRepository) FindAll(ctx context.Context, filter model.ArticleQuery) ([]*model.Article, int, error) {
	var (
		args       []any
		conditions []string
	)

	cacheableLimit := filter.Limit
	if cacheableLimit <= 0 {
		cacheableLimit = model.CacheableLimit
	}

	shouldCache := filter.Query == "" && filter.Page == 1 && cacheableLimit == model.CacheableLimit
	var cacheKey string
	if shouldCache {
		cacheKey := fmt.Sprintf("%s:page=%d:limit=%d", model.ArticleKey, filter.Page, filter.Limit)

		var cached struct {
			Results []*model.Article
			Total   int
		}
		if err := r.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached.Results, cached.Total, nil
		}
	}

	baseQuery := `
		SELECT a.id, a.author_id, au.name, a.title, a.body, a.created_at
		FROM articles a
		JOIN authors au ON a.author_id = au.id
	`

	argPos := 1
	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(a.title ILIKE $%d OR a.body ILIKE $%d)", argPos, argPos+1))
		args = append(args, "%"+filter.Query+"%", "%"+filter.Query+"%")
		argPos += 2
	}
	if filter.Author != "" {
		conditions = append(conditions, fmt.Sprintf("au.name ILIKE $%d", argPos))
		args = append(args, "%"+filter.Author+"%")
		argPos++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := (filter.Page - 1) * limit
	if filter.Page <= 0 {
		offset = 0
	}

	args = append(args, limit, offset)
	limitPos := argPos
	offsetPos := argPos + 1

	fullQuery := fmt.Sprintf(
		"%s%s ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d",
		baseQuery,
		whereClause,
		limitPos,
		offsetPos,
	)

	baseCount := `
		SELECT COUNT(*)
		FROM articles a
		JOIN authors au ON a.author_id = au.id
	`
	countQuery := baseCount + whereClause

	rows, err := r.db.QueryContext(ctx, fullQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*model.Article
	for rows.Next() {
		var a model.Article
		if err := rows.Scan(&a.ID, &a.AuthorID, &a.Author, &a.Title, &a.Body, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, &a)
	}

	countArgs := args[:len(args)-2]
	var total int
	err = r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if shouldCache {
		_ = r.cache.Set(ctx, cacheKey, struct {
			Results []*model.Article
			Total   int
		}{results, total}, config.RedisExpired())
	}

	return results, total, nil
}

func (r *articleRepository) Create(ctx context.Context, article *model.Article) (*model.Article, error) {
	article.ID = uuid.New()

	query := `
		INSERT INTO articles (id, author_id, title, body, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		article.ID,
		article.AuthorID,
		article.Title,
		article.Body,
	).Scan(&article.CreatedAt)
	if err != nil {
		return nil, err
	}

	_ = r.cache.Delete(ctx, fmt.Sprintf("%s:page=1:limit=%d", model.ArticleKey, model.CacheableLimit))

	return article, nil
}
