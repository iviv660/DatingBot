package usecase

import (
	"app/match/internal/dto"
	"context"
)

type MatchRepo interface {
	Like(ctx context.Context, fromUser, toUser int64, isLike bool) error
	CheckMatch(ctx context.Context, user1, user2 int64) (bool, error)
	TodayLikedIDs(ctx context.Context, fromUser int64) ([]int64, error)
}

type UserClient interface {
	GetByTelegramID(ctx context.Context, telegramID int64) (*dto.User, error)
	GetCandidates(context.Context, dto.Candidate) ([]*dto.User, error)
}
