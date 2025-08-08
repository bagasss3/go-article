package model

import (
	"context"
	"time"

	"github.com/google/uuid"
)

var (
	ArticleKey     string = "article"
	CacheableLimit int    = 10
)

type ArticleQuery struct {
	Query  string `query:"query"`
	Author string `query:"author"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type Article struct {
	ID        uuid.UUID `json:"id"`
	AuthorID  uuid.UUID `json:"author_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`

	Author string `json:"author"`
}

type CachedArticles struct {
	Results []*Article `json:"results"`
	Total   int        `json:"total"`
}

type CreateArticleRequest struct {
	AuthorID string `json:"author_id" validate:"required,uuid"`
	Title    string `json:"title" validate:"required,min=3,max=255"`
	Body     string `json:"body" validate:"required"`
}

type ArticleMethodService interface {
	FindAll(ctx context.Context, filter ArticleQuery) ([]*Article, int, error)
	Create(ctx context.Context, req *CreateArticleRequest) (*Article, error)
}

type ArticleRepository interface {
	FindAll(ctx context.Context, filter ArticleQuery) ([]*Article, int, error)
	Create(ctx context.Context, article *Article) (*Article, error)
}
