package client

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

	userpb "app/user/proto"
)

var ErrEmptyResponse = errors.New("user service returned empty response")

type UserClientAdapter struct {
	grpc userpb.UserServiceClient
}

func NewUserClientAdapter(grpc userpb.UserServiceClient) *UserClientAdapter {
	return &UserClientAdapter{grpc: grpc}
}

func (c *UserClientAdapter) GetByID(ctx context.Context, id int64) (*userpb.User, error) {
	resp, err := c.grpc.GetProfile(ctx, &userpb.GetProfileRequest{UserId: id})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.User == nil {
		return nil, ErrEmptyResponse
	}
	return resp.User, nil
}

func (c *UserClientAdapter) GetByTelegramID(ctx context.Context, telegramID int64) (*userpb.User, error) {
	resp, err := c.grpc.GetByTelegramID(ctx, &userpb.GetByTelegramIDRequest{TelegramId: telegramID})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.User == nil {
		return nil, ErrEmptyResponse
	}
	return resp.User, nil
}

func (c *UserClientAdapter) Create(ctx context.Context, u *userpb.User) (*userpb.User, error) {
	if u == nil {
		return nil, errors.New("nil user")
	}
	req := &userpb.RegisterUserRequest{
		TelegramId:  u.GetTelegramId(),
		Username:    u.GetUsername(),
		Age:         u.GetAge(),
		Gender:      u.GetGender(),
		Location:    u.GetLocation(),
		Description: u.GetDescription(),
		IsVisible:   u.GetIsVisible(),
	}
	resp, err := c.grpc.RegisterUser(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.User == nil {
		return nil, ErrEmptyResponse
	}
	return resp.User, nil
}

func (c *UserClientAdapter) Update(ctx context.Context, u *userpb.User) (*userpb.User, error) {
	if u == nil {
		return nil, errors.New("nil user")
	}
	req := &userpb.UpdateProfileRequest{
		UserId:      u.GetId(),
		Username:    u.GetUsername(),
		Age:         u.GetAge(),
		Gender:      u.GetGender(),
		Location:    u.GetLocation(),
		Description: u.GetDescription(),
		IsVisible:   u.GetIsVisible(),
	}
	resp, err := c.grpc.UpdateProfile(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.User == nil {
		return nil, ErrEmptyResponse
	}
	return resp.User, nil
}

func (c *UserClientAdapter) UpdatePhoto(ctx context.Context, userID int64, photo io.Reader) (*userpb.User, error) {
	if photo == nil {
		return nil, errors.New("nil photo reader")
	}
	data, err := ioutil.ReadAll(photo)
	if err != nil {
		return nil, err
	}
	req := &userpb.PhotoUploadRequest{
		UserId: userID,
		File:   data,
	}
	_, err = c.grpc.PhotoUpload(ctx, req)
	if err != nil {
		return nil, err
	}
	return c.GetByID(ctx, userID)
}

func (c *UserClientAdapter) ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error {
	req := &userpb.ToggleVisibilityRequest{
		UserId:    userID,
		IsVisible: isVisible,
	}
	resp, err := c.grpc.ToggleVisibility(ctx, req)
	if err != nil {
		return err
	}
	if resp == nil {
		return ErrEmptyResponse
	}
	return nil
}
