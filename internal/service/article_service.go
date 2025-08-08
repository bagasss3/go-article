package service

import (
	"context"

	"github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/pkg/model"

	"github.com/google/uuid"
)

type articleService struct {
	articleRepository model.ArticleRepository
	authorRepository  model.AuthorRepository
}

func NewArticleService(articleRepository model.ArticleRepository, authorRepository model.AuthorRepository) model.ArticleMethodService {
	return &articleService{
		articleRepository: articleRepository,
		authorRepository:  authorRepository,
	}
}

func (s *articleService) FindAll(ctx context.Context, filter model.ArticleQuery) ([]*model.Article, int, error) {
	articles, total, err := s.articleRepository.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if len(articles) <= 0 {
		return []*model.Article{}, 0, nil
	}

	return articles, total, nil
}

func (s *articleService) Create(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error) {
	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		return nil, errors.New(errors.ErrInvalidData, "invalid author id format")
	}

	author, err := s.authorRepository.FindByID(ctx, authorID)
	if err != nil {
		return nil, err
	}

	if author == nil {
		return nil, errors.New(errors.ErrRecordNotFound, "author not found")
	}

	article := &model.Article{
		ID:       uuid.New(),
		AuthorID: authorID,
		Title:    req.Title,
		Body:     req.Body,
	}

	return s.articleRepository.Create(ctx, article)
}
