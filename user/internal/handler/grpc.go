package handler

import (
	"bytes"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"

	"app/user/internal/dto"
	"app/user/internal/entity"
	"app/user/internal/usecase"
	userpb "app/user/proto"
)

type Handler struct {
	userpb.UnimplementedUserServiceServer
	uc *usecase.Usecase
}

func NewHandler(uc *usecase.Usecase) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) GetByTelegramID(ctx context.Context, req *userpb.GetByTelegramIDRequest) (*userpb.UserResponse, error) {
	u, err := h.uc.GetUserByTelegramID(ctx, req.GetTelegramId())
	if err != nil {
		return nil, err
	}
	return &userpb.UserResponse{User: toPB(u)}, nil
}

func (h *Handler) RegisterUser(ctx context.Context, req *userpb.RegisterUserRequest) (*userpb.UserResponse, error) {
	u := &entity.User{
		TelegramID:  req.GetTelegramId(),
		Username:    req.GetUsername(),
		Age:         int(req.GetAge()),
		Gender:      req.GetGender(),
		Location:    req.GetLocation(),
		Description: req.GetDescription(),
		IsVisible:   req.GetIsVisible(),
	}
	created, err := h.uc.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	return &userpb.UserResponse{User: toPB(created)}, nil
}

func (h *Handler) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.UserResponse, error) {
	u, err := h.uc.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}
	return &userpb.UserResponse{User: toPB(u)}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UserResponse, error) {
	u := &entity.User{
		ID:          req.GetUserId(),
		Username:    req.GetUsername(),
		Age:         int(req.GetAge()),
		Gender:      req.GetGender(),
		Location:    req.GetLocation(),
		Description: req.GetDescription(),
		IsVisible:   req.GetIsVisible(),
	}
	updated, err := h.uc.Update(ctx, u)
	if err != nil {
		return nil, err
	}
	return &userpb.UserResponse{User: toPB(updated)}, nil
}

func (h *Handler) GetCandidates(ctx context.Context, req *userpb.GetCandidatesRequest) (*userpb.GetCandidatesResponse, error) {
	filter := dto.CandidateFilter{
		TargetGender: req.GetTargetGender(),
		MinAge:       int(req.GetMinAge()),
		MaxAge:       int(req.GetMaxAge()),
		Location:     req.GetLocation(),
		Limit:        int(req.GetLimit()),
	}
	list, err := h.uc.GetCandidatProfiles(ctx, filter)
	if err != nil {
		return nil, err
	}
	out := make([]*userpb.User, 0, len(list))
	for _, u := range list {
		out = append(out, toPB(u))
	}
	return &userpb.GetCandidatesResponse{Candidates: out}, nil
}

func (h *Handler) ToggleVisibility(ctx context.Context, req *userpb.ToggleVisibilityRequest) (*userpb.ToggleVisibilityResponse, error) {
	if err := h.uc.ToggleVisibility(ctx, req.GetUserId(), req.GetIsVisible()); err != nil {
		return nil, err
	}
	return &userpb.ToggleVisibilityResponse{Success: true}, nil
}

func (h *Handler) PhotoUpload(ctx context.Context, req *userpb.PhotoUploadRequest) (*userpb.PhotoUploadResponse, error) {
	if req == nil || len(req.GetFile()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty file")
	}

	if _, err := h.uc.UploadPhoto(ctx, req.GetUserId(), bytes.NewReader(req.GetFile())); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	u, err := h.uc.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &userpb.PhotoUploadResponse{PhotoUrl: u.PhotoURL}, nil
}

// --- helpers ---

func toPB(u *entity.User) *userpb.User {
	if u == nil {
		return nil
	}
	return &userpb.User{
		Id:          u.ID,
		TelegramId:  u.TelegramID,
		Username:    u.Username,
		Age:         int32(u.Age),
		Gender:      u.Gender,
		Location:    u.Location,
		Description: u.Description,
		PhotoUrl:    u.PhotoURL,
		IsVisible:   u.IsVisible,
		CreatedAt:   u.CreatedAt.Format(time.RFC3339),
	}
}
