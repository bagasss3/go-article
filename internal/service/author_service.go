package service

import (
	"context"

	"github.com/bagasss3/go-article/internal/errors"

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
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, errors.New(errors.ErrInvalidData, "invalid author ID format")
	}

	author, err := s.authorRepository.FindByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if author == nil {
		return nil, errors.New(errors.ErrRecordNotFound, "author not found")
	}

	return author, nil
}

func (s *authorService) Create(ctx context.Context, req *model.CreateAuthorRequest) (*model.Author, error) {
	author := &model.Author{
		Name: req.Name,
	}

	return s.authorRepository.Create(ctx, author)
}
