package handler

import (
	"context"

	"app/match/internal/usecase"
	matchpb "app/match/proto"
)

type Handler struct {
	matchpb.UnimplementedMatchServiceServer
	uc *usecase.Usecase
}

func NewHandler(uc *usecase.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) Like(ctx context.Context, req *matchpb.LikeRequest) (*matchpb.LikeResponse, error) {
	if err := h.uc.Like(ctx, req.GetFromUser(), req.GetToUser(), req.GetIsLike()); err != nil {
		return nil, err
	}
	return &matchpb.LikeResponse{Success: true}, nil
}

func (h *Handler) CheckMatch(ctx context.Context, req *matchpb.CheckMatchRequest) (*matchpb.CheckMatchResponse, error) {
	ok, err := h.uc.Match(ctx, req.GetUser1(), req.GetUser2())
	if err != nil {
		return nil, err
	}
	return &matchpb.CheckMatchResponse{Match: ok}, nil
}

func (h *Handler) GetCandidates(ctx context.Context, req *matchpb.GetCandidatesRequest) (*matchpb.GetCandidatesResponse, error) {
	list, err := h.uc.GetCandidats(ctx, req.GetTelegramId())
	if err != nil {
		return nil, err
	}

	out := make([]*matchpb.User, 0, len(list))
	for _, u := range list {
		out = append(out, &matchpb.User{
			Id:          u.ID,
			TelegramId:  u.TelegramID,
			Username:    u.Username,
			Age:         int32(u.Age),
			Gender:      u.Gender,
			Location:    u.Location,
			Description: u.Description,
			PhotoUrl:    u.PhotoURL,
			IsVisible:   u.IsVisible,
		})
	}

	return &matchpb.GetCandidatesResponse{Candidates: out}, nil
}
