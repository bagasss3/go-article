package service

import (
	"context"

	"github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/internal/helper"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
)

type articleService struct {
	articleRepository model.ArticleRepository
}

func NewArticleService(articleRepository model.ArticleRepository) model.ArticleMethodService {
	return &articleService{
		articleRepository: articleRepository,
	}
}

func (s *articleService) FindAll(ctx context.Context, filter model.ArticleQuery) ([]*model.Article, int, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	articles, total, err := s.articleRepository.FindAll(ctx, filter)
	if err != nil {
		log.Error(err)
		return nil, 0, err
	}

	if len(articles) <= 0 {
		return []*model.Article{}, 0, nil
	}

	return articles, total, nil
}

func (s *articleService) Create(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error) {
	log := logrus.WithFields(logrus.Fields{
		"req": helper.ToJSON(req),
	})

	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		err := errors.New(errors.ErrInvalidData, "invalid author id format")
		log.Error(err)
		return nil, err
	}

	// author, err := s.authorRepository.FindByID(ctx, authorID)
	// if err != nil {
	// 	log.Error(err)
	// 	return nil, err
	// }

	// if author == nil {
	// 	err := errors.New(errors.ErrRecordNotFound, "author not found")
	// 	log.Error(err)
	// 	return nil, err
	// }

	article := &model.Article{
		ID:       uuid.New(),
		AuthorID: authorID,
		Title:    req.Title,
		Body:     req.Body,
	}

	result, err := s.articleRepository.Create(ctx, article)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return result, nil
}
