package mocks

import (
	"context"
	"io"

	"github.com/stretchr/testify/mock"
)

type MockMinioRepository struct {
	mock.Mock
}

func NewMockMinioRepository() *MockMinioRepository {
	return &MockMinioRepository{}
}

func (m *MockMinioRepository) Upload(ctx context.Context, userID int64, file io.Reader) (string, error) {
	args := m.Called(ctx, userID, file)
	return args.String(0), args.Error(1)
}
