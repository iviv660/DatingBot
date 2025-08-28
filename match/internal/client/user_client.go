package client

import (
	"app/match/internal/dto"
	"context"

	userpb "app/user/proto"
)

type UserClientAdapter struct {
	grpc userpb.UserServiceClient
}

func NewUserClientAdapter(grpc userpb.UserServiceClient) *UserClientAdapter {
	return &UserClientAdapter{grpc: grpc}
}

func (c *UserClientAdapter) GetProfile(ctx context.Context, userID int64) (*dto.User, error) {
	resp, err := c.grpc.GetProfile(ctx, &userpb.GetProfileRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return fromPB(resp.User), nil
}

func (c *UserClientAdapter) GetByTelegramID(ctx context.Context, telegramID int64) (*dto.User, error) {
	resp, err := c.grpc.GetByTelegramID(ctx, &userpb.GetByTelegramIDRequest{TelegramId: telegramID})
	if err != nil {
		return nil, err
	}
	return fromPB(resp.User), nil
}

func (c *UserClientAdapter) GetCandidates(ctx context.Context, cand dto.Candidate) ([]*dto.User, error) {
	resp, err := c.grpc.GetCandidates(ctx, &userpb.GetCandidatesRequest{
		TargetGender: cand.TargetGender,
		MinAge:       int32(cand.MinAge),
		MaxAge:       int32(cand.MaxAge),
		Location:     cand.Location,
		Limit:        int32(cand.Limit),
	})
	if err != nil {
		return nil, err
	}
	users := make([]*dto.User, 0, len(resp.Candidates))
	for _, u := range resp.Candidates {
		users = append(users, fromPB(u))
	}
	return users, nil
}

func fromPB(u *userpb.User) *dto.User {
	if u == nil {
		return nil
	}
	return &dto.User{
		ID:          u.Id,
		TelegramID:  u.TelegramId,
		Username:    u.Username,
		Age:         int(u.Age),
		Gender:      u.Gender,
		Location:    u.Location,
		Description: u.Description,
		PhotoURL:    u.PhotoUrl,
		IsVisible:   u.IsVisible,
	}
}
