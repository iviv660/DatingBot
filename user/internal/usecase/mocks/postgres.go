package mocks

import (
	"app/user/internal/dto"
	"app/user/internal/entity"
	"context"

	"github.com/stretchr/testify/mock"
)

type MockPostgresRepository struct {
	mock.Mock
}

func NewMockPostgresRepository() *MockPostgresRepository {
	return &MockPostgresRepository{}
}

func (m *MockPostgresRepository) GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	args := m.Called(ctx, telegramID)
	if args.Get(0) != nil {
		return args.Get(0).(*entity.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPostgresRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockPostgresRepository) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *MockPostgresRepository) UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) (*entity.User, error) {
	args := m.Called(ctx, userID, input)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockPostgresRepository) GetCandidates(ctx context.Context, filter dto.CandidateFilter) ([]*entity.User, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*entity.User), args.Error(1)
}

func (m *MockPostgresRepository) ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error {
	args := m.Called(ctx, userID, isVisible)
	return args.Error(0)
}

func (m *MockPostgresRepository) UpdatePhoto(ctx context.Context, userID int64, url string) error {
	args := m.Called(ctx, userID, url)
	return args.Error(0)
}
