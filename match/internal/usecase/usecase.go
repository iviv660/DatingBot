package usecase

import (
	"app/match/internal/dto"
	"app/match/internal/utils"
	"context"
)

type Usecase struct {
	repo       MatchRepo
	userClient UserClient
}

func NewUseCase(repo MatchRepo, userClient UserClient) *Usecase {
	return &Usecase{repo: repo, userClient: userClient}
}

func (u *Usecase) Like(ctx context.Context, fromUser, toUser int64, isLike bool) error {
	return u.repo.Like(ctx, fromUser, toUser, isLike)
}

func (u *Usecase) Match(ctx context.Context, fromUser int64, toUser int64) (bool, error) {
	return u.repo.CheckMatch(ctx, fromUser, toUser)
}

func (u *Usecase) GetCandidats(ctx context.Context, telegramID int64) ([]*dto.User, error) {
	me, err := u.userClient.GetByTelegramID(ctx, telegramID)
	if err != nil {
		return nil, err
	}

	exclude, err := u.repo.TodayLikedIDs(ctx, me.ID)
	if err != nil {
		return nil, err
	}

	filter := dto.Candidate{
		TargetGender: utils.OppositeGender(me.Gender),
		MinAge:       me.Age - 3,
		MaxAge:       me.Age + 3,
		Location:     me.Location,
		Limit:        20,
		ExcludeIDs:   exclude,
	}

	list, err := u.userClient.GetCandidates(ctx, filter)
	if err != nil {
		return nil, err
	}
	return list, nil
}
