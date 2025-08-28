package internal

import (
	matchpb "app/match/proto"

	userpb "app/user/proto"
	"context"
	"io"
)

type UserClient interface {
	GetByID(ctx context.Context, id int64) (*userpb.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*userpb.User, error)
	Create(ctx context.Context, user *userpb.User) (*userpb.User, error)
	Update(ctx context.Context, user *userpb.User) (*userpb.User, error)
	UpdatePhoto(ctx context.Context, userID int64, photo io.Reader) (*userpb.User, error)
	ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error
}

type MatchClient interface {
	GetCandidates(ctx context.Context, userID int64) ([]*matchpb.User, error)
	Like(ctx context.Context, fromUserID int64, toUserID int64, isLike bool) error
	Match(ctx context.Context, fromUserID, toUserId int64) (bool, error)
}
