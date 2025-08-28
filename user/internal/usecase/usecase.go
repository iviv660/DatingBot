package usecase

import (
	"app/user/internal/dto"
	"app/user/internal/entity"
	"context"
	"errors"
	"io"
	"log"
)

type Usecase struct {
	repo     Repo
	cache    Cache
	uploader PhotoUploader
}

func New(repo Repo, cache Cache, uploader PhotoUploader) *Usecase {
	return &Usecase{repo: repo, cache: cache, uploader: uploader}
}

func (uc *Usecase) GetUserByTelegramID(ctx context.Context, telegramID int64) (*entity.User, error) {
	user, err := uc.repo.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *Usecase) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	user, err := uc.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	if err = uc.cache.SetProfile(ctx, user); err != nil {
		log.Println(err)
	}
	return user, nil
}

func (uc *Usecase) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	user, err := uc.cache.GetProfile(ctx, id)
	if err != nil {
		log.Println(err)
	}
	if user != nil {
		return user, nil
	}
	user, err = uc.repo.GetProfile(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = uc.cache.SetProfile(ctx, user); err != nil {
		log.Println(err)
	}
	return user, nil
}

func (uc *Usecase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	input := dto.UpdateProfileInput{
		Username:    user.Username,
		Age:         user.Age,
		Gender:      user.Gender,
		Location:    user.Location,
		Description: user.Description,
		IsVisible:   user.IsVisible,
	}

	updatedUser, err := uc.repo.UpdateProfile(ctx, user.ID, input)
	if err != nil {
		return nil, err
	}

	if err := uc.cache.Invalidate(ctx, user.ID); err != nil {
		log.Println("cache invalidate error:", err)
	}

	return updatedUser, nil
}

func (uc *Usecase) GetCandidatProfiles(ctx context.Context, filter dto.CandidateFilter) ([]*entity.User, error) {
	candidates, err := uc.repo.GetCandidates(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, errors.New("no candidates found")
	}

	return candidates, nil
}

func (uc *Usecase) ToggleVisibility(ctx context.Context, userID int64, isVisible bool) error {
	if err := uc.repo.ToggleVisibility(ctx, userID, isVisible); err != nil {
		return err
	}

	if err := uc.cache.Invalidate(ctx, userID); err != nil {
		log.Println("cache invalidate error:", err)
	}

	return nil
}

func (uc *Usecase) UploadPhoto(ctx context.Context, userID int64, file io.Reader) (string, error) {
	url, err := uc.uploader.Upload(ctx, userID, file)
	if err != nil {
		return "", err
	}

	if err := uc.repo.UpdatePhoto(ctx, userID, url); err != nil {
		return "", err
	}

	if err := uc.cache.Invalidate(ctx, userID); err != nil {
		log.Println("cache invalidate error:", err)
	}
	return url, nil
}
