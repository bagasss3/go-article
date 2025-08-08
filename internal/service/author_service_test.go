package service

import (
	"context"
	"errors"
	"testing"

	customErrors "github.com/bagasss3/go-article/internal/errors"
	"github.com/bagasss3/go-article/internal/mocks"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewAuthorService(t *testing.T) {
	s := NewAuthorService(nil)
	require.NotNil(t, s)
}

func TestAuthorService_FindByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockRepo := mocks.NewMockAuthorRepository(ctrl)

	service := &authorService{authorRepository: mockRepo}

	t.Run("invalid uuid", func(t *testing.T) {
		res, err := service.FindByID(ctx, "not-a-uuid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), customErrors.ErrInvalidData.Error())
		assert.Nil(t, res)
	})

	t.Run("author not found", func(t *testing.T) {
		uid := uuid.New()

		mockRepo.EXPECT().
			FindByID(gomock.Any(), uid).
			Return(nil, nil)

		res, err := service.FindByID(ctx, uid.String())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), customErrors.ErrRecordNotFound.Error())
		assert.Nil(t, res)
	})

	t.Run("repo error", func(t *testing.T) {
		uid := uuid.New()

		mockRepo.EXPECT().
			FindByID(gomock.Any(), uid).
			Return(nil, errors.New("db failure"))

		res, err := service.FindByID(ctx, uid.String())
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		uid := uuid.New()
		expected := &model.Author{ID: uid, Name: "Test Author"}

		mockRepo.EXPECT().
			FindByID(gomock.Any(), uid).
			Return(expected, nil)

		res, err := service.FindByID(ctx, uid.String())
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}

func TestAuthorService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.TODO()
	mockRepo := mocks.NewMockAuthorRepository(ctrl)

	service := &authorService{authorRepository: mockRepo}

	t.Run("repo error", func(t *testing.T) {
		req := &model.CreateAuthorRequest{Name: "Failing Author"}

		mockRepo.EXPECT().
			Create(gomock.Any(), &model.Author{Name: req.Name}).
			Return(nil, errors.New("insert error"))

		res, err := service.Create(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("success", func(t *testing.T) {
		req := &model.CreateAuthorRequest{Name: "John Doe"}
		expected := &model.Author{ID: uuid.New(), Name: req.Name}

		mockRepo.EXPECT().
			Create(gomock.Any(), &model.Author{Name: req.Name}).
			Return(expected, nil)

		res, err := service.Create(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
	})
}
