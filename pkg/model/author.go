package model

import (
	"context"

	"github.com/google/uuid"
)

var (
	AuthorKey string = "author"
)

type Author struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CreateAuthorRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}

type AuthorRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Author, error)
	Create(ctx context.Context, author *Author) (*Author, error)
}

type AuthorMethodService interface {
	FindByID(ctx context.Context, id string) (*Author, error)
	Create(ctx context.Context, req *CreateAuthorRequest) (*Author, error)
}
