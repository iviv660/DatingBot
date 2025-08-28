package client

import (
	"context"
	"errors"

	matchpb "app/match/proto"
)

var ErrMatchEmptyResponse = errors.New("match service returned empty response")

type MatchClientAdapter struct {
	grpc matchpb.MatchServiceClient
}

func NewMatchClientAdapter(grpc matchpb.MatchServiceClient) *MatchClientAdapter {
	return &MatchClientAdapter{grpc: grpc}
}

func (c *MatchClientAdapter) GetCandidates(ctx context.Context, telegramID int64) ([]*matchpb.User, error) {
	resp, err := c.grpc.GetCandidates(ctx, &matchpb.GetCandidatesRequest{
		TelegramId: telegramID,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Candidates == nil {
		return []*matchpb.User{}, nil
	}
	return resp.Candidates, nil
}

func (c *MatchClientAdapter) Like(ctx context.Context, fromUserID, toUserID int64, isLike bool) error {
	resp, err := c.grpc.Like(ctx, &matchpb.LikeRequest{
		FromUser: fromUserID,
		ToUser:   toUserID,
		IsLike:   isLike,
	})
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrMatchEmptyResponse
	}
	if !resp.Success {
		return errors.New("like not processed")
	}
	return nil
}

func (c *MatchClientAdapter) Match(ctx context.Context, fromUserID, toUserID int64) (bool, error) {
	resp, err := c.grpc.CheckMatch(ctx, &matchpb.CheckMatchRequest{
		User1: fromUserID,
		User2: toUserID,
	})
	if err != nil {
		return false, err
	}
	if resp == nil {
		return false, ErrMatchEmptyResponse
	}
	return resp.Match, nil
}
