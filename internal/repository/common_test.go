package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/stretchr/testify/require"
)

type repoTestKit struct {
	db        *sql.DB
	mock      sqlmock.Sqlmock
	cache     cache.Cache
	mockCache *cache.MockCache
	closer    func()
}

func initializeRepoTestKit(t *testing.T) *repoTestKit {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	mockCache := cache.NewMockCache()

	return &repoTestKit{
		db:        db,
		mock:      mock,
		cache:     mockCache,
		mockCache: mockCache,
		closer: func() {
			_ = db.Close()
		},
	}
}
