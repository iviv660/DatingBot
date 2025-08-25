package usecase

import (
	"app/user/internal/dto"
	"app/user/internal/entity"
	"context"
	"io"
)

type UserRepo interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetProfile(ctx context.Context, userID int64) (*entity.User, error)
	UpdateProfile(ctx context.Context, userID int64, input dto.UpdateProfileInput) (*entity.User, error)
	GetCandidates(ctx context.Context, filter dto.CandidateFilter) ([]*entity.User, error)
	ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error
}

type CacheRepo interface {
	SetProfile(ctx context.Context, user *entity.User) error
	GetProfile(ctx context.Context, userID int64) (*entity.User, error)
	Invalidate(ctx context.Context, userID int64) error
}

type PhotoUploader interface {
	Upload(ctx context.Context, userID int64, file io.Reader) (string, error)
}
