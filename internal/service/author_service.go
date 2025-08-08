package service

import (
	"context"

	"github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/internal/helper"
	"github.com/sirupsen/logrus"

	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
)

type authorService struct {
	authorRepository model.AuthorRepository
}

func NewAuthorService(authorRepository model.AuthorRepository) model.AuthorMethodService {
	return &authorService{
		authorRepository: authorRepository,
	}
}

func (s *authorService) FindByID(ctx context.Context, id string) (*model.Author, error) {
	log := logrus.WithFields(logrus.Fields{
		"author_id": id,
	})

	uid, err := uuid.Parse(id)
	if err != nil {
		err := errors.New(errors.ErrInvalidData, "invalid author ID format")
		log.Error(err)
		return nil, err
	}

	author, err := s.authorRepository.FindByID(ctx, uid)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if author == nil {
		err := errors.New(errors.ErrRecordNotFound, "author not found")
		log.Error(err)
		return nil, err
	}

	return author, nil
}

func (s *authorService) Create(ctx context.Context, req *model.CreateAuthorRequest) (*model.Author, error) {
	log := logrus.WithFields(logrus.Fields{
		"req": helper.ToJSON(req),
	})

	author := &model.Author{
		Name: req.Name,
	}

	result, err := s.authorRepository.Create(ctx, author)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return result, nil
}
