package mocks

import (
	"app/user/internal/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockRedisRepository struct {
	mock.Mock
}

func NewMockRedisRepository() *MockRedisRepository {
	return &MockRedisRepository{}
}

func (r *MockRedisRepository) SetProfile(ctx context.Context, user *entity.User) error {
	args := r.Called(ctx, user)
	return args.Error(0)
}

func (r *MockRedisRepository) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	args := r.Called(ctx, userID)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (r *MockRedisRepository) Invalidate(ctx context.Context, userID int64) error {
	args := r.Called(ctx, userID)
	return args.Error(0)
}
