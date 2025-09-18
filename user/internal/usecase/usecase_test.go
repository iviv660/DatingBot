package usecase

import (
	"app/user/internal/dto"
	"app/user/internal/entity"
	"app/user/internal/usecase/mocks"
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func UCInit() (*Usecase, *mocks.MockPostgresRepository, *mocks.MockRedisRepository, *mocks.MockMinioRepository) {
	pg := mocks.NewMockPostgresRepository()
	redis := mocks.NewMockRedisRepository()
	minio := mocks.NewMockMinioRepository()
	uc := New(pg, redis, minio)
	return uc, pg, redis, minio
}

func TestUseCase_GetUserByTelegramID(t *testing.T) {
	uc, pg, _, _ := UCInit()

	fixedTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	expected := &entity.User{
		ID:          1,
		TelegramID:  42,
		Username:    "Volodya",
		Age:         25,
		Gender:      "male",
		Location:    "Vladivostok",
		Description: "Backend developer, loves Go",
		PhotoURL:    "https://example.com/photos/42.jpg",
		CreatedAt:   fixedTime,
		IsVisible:   true,
	}

	// happy-path
	pg.On("GetByTelegramID", mock.Anything, int64(42)).
		Return(expected, nil)

	user, err := uc.GetUserByTelegramID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(user, expected) {
		t.Errorf("got %+v, want %+v", user, expected)
	}

	pg.AssertExpectations(t)
}

func TestUseCase_Create(t *testing.T) {
	uc, pg, redis, _ := UCInit()

	expected := &entity.User{
		ID:          1,
		TelegramID:  42,
		Username:    "Volodya",
		Age:         25,
		Gender:      "male",
		Location:    "Vladivostok",
		Description: "Backend developer, loves Go",
		PhotoURL:    "https://example.com/photos/42.jpg",
		CreatedAt:   time.Now(), // в реальных тестах лучше зафиксировать время
		IsVisible:   true,
	}

	pg.On("Create", mock.Anything, mock.Anything).
		Return(expected, nil)

	redis.On("SetProfile", mock.Anything, mock.Anything).
		Return(nil)

	user, err := uc.Create(context.Background(), expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(user, expected) {
		t.Errorf("got %+v, want %+v", user, expected)
	}
	pg.AssertExpectations(t)
	redis.AssertExpectations(t)
}

func TestUseCase_GetUserByID(t *testing.T) {
	uc, pg, redis, _ := UCInit()

	fixedTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	expected := &entity.User{
		ID:          1,
		TelegramID:  42,
		Username:    "Volodya",
		Age:         25,
		Gender:      "male",
		Location:    "Vladivostok",
		Description: "Backend developer, loves Go",
		PhotoURL:    "https://example.com/photos/42.jpg",
		CreatedAt:   fixedTime,
		IsVisible:   true,
	}

	// 1. Кеш пустой → возвращаем nil
	redis.On("GetProfile", mock.Anything, int64(1)).
		Return((*entity.User)(nil), nil)

	// 2. В PG есть юзер
	pg.On("GetProfile", mock.Anything, int64(1)).
		Return(expected, nil)

	// 3. После получения из PG, юзер кладётся в кеш
	redis.On("SetProfile", mock.Anything, expected).
		Return(nil)

	user, err := uc.GetUserByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(user, expected) {
		t.Errorf("got %+v, want %+v", user, expected)
	}

	pg.AssertExpectations(t)
	redis.AssertExpectations(t)
}

func TestUseCase_Update(t *testing.T) {
	uc, pg, redis, _ := UCInit()

	fixedTime := time.Date(2025, 9, 17, 12, 0, 0, 0, time.UTC)
	expected := &entity.User{
		ID:          1,
		TelegramID:  42,
		Username:    "Volodya",
		Age:         25,
		Gender:      "male",
		Location:    "Vladivostok",
		Description: "Backend developer, loves Go",
		PhotoURL:    "https://example.com/photos/42.jpg",
		CreatedAt:   fixedTime,
		IsVisible:   true,
	}

	// мокаем Postgres UpdateProfile
	pg.On("UpdateProfile", mock.Anything, expected.ID, mock.Anything).
		Return(expected, nil)

	// мокаем Redis Invalidate
	redis.On("Invalidate", mock.Anything, expected.ID).
		Return(nil)

	user, err := uc.Update(context.Background(), expected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(user, expected) {
		t.Errorf("got %+v, want %+v", user, expected)
	}

	pg.AssertExpectations(t)
	redis.AssertExpectations(t)
}

func TestUseCase_GetCandidatProfiles(t *testing.T) {
	uc, pg, _, _ := UCInit()

	filter := dto.CandidateFilter{
		TargetGender: "female",
		MinAge:       20,
		MaxAge:       30,
		Location:     "Vladivostok",
		Limit:        5,
		ExcludeIDs:   []int64{42},
	}

	expected := []*entity.User{
		{
			ID:         1,
			TelegramID: 1001,
			Username:   "Anna",
			Age:        25,
			Gender:     "female",
			Location:   "Vladivostok",
			IsVisible:  true,
		},
		{
			ID:         2,
			TelegramID: 1002,
			Username:   "Maria",
			Age:        27,
			Gender:     "female",
			Location:   "Vladivostok",
			IsVisible:  true,
		},
	}

	// мокаем PG
	pg.On("GetCandidates", mock.Anything, filter).
		Return(expected, nil)

	candidates, err := uc.GetCandidatProfiles(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(candidates, expected) {
		t.Errorf("got %+v, want %+v", candidates, expected)
	}

	pg.AssertExpectations(t)
}

func TestUseCase_ToggleVisibility(t *testing.T) {
	uc, pg, redis, _ := UCInit()

	tests := []struct {
		name      string
		repoErr   error
		cacheErr  error
		expectErr bool
	}{
		{
			name:      "happy-path",
			repoErr:   nil,
			cacheErr:  nil,
			expectErr: false,
		},
		{
			name:      "repo error",
			repoErr:   errors.New("db error"),
			cacheErr:  nil,
			expectErr: true,
		},
		{
			name:      "cache error is ignored",
			repoErr:   nil,
			cacheErr:  errors.New("redis down"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg.ExpectedCalls = nil
			redis.ExpectedCalls = nil

			pg.On("ToggleVisibility", mock.Anything, int64(1), true).
				Return(tt.repoErr)

			if tt.repoErr == nil {
				redis.On("Invalidate", mock.Anything, int64(1)).
					Return(tt.cacheErr)
			}

			err := uc.ToggleVisibility(context.Background(), 1, true)
			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			pg.AssertExpectations(t)
			redis.AssertExpectations(t)
		})
	}
}

func TestUseCase_UploadPhoto(t *testing.T) {
	uc, pg, redis, minio := UCInit()

	tests := []struct {
		name      string
		uploadURL string
		uploadErr error
		repoErr   error
		cacheErr  error
		expectErr bool
	}{
		{
			name:      "happy-path",
			uploadURL: "http://cdn/pic.jpg",
			uploadErr: nil,
			repoErr:   nil,
			cacheErr:  nil,
			expectErr: false,
		},
		{
			name:      "uploader error",
			uploadURL: "",
			uploadErr: errors.New("upload failed"),
			repoErr:   nil,
			cacheErr:  nil,
			expectErr: true,
		},
		{
			name:      "repo error",
			uploadURL: "http://cdn/pic.jpg",
			uploadErr: nil,
			repoErr:   errors.New("db error"),
			cacheErr:  nil,
			expectErr: true,
		},
		{
			name:      "cache error ignored",
			uploadURL: "http://cdn/pic.jpg",
			uploadErr: nil,
			repoErr:   nil,
			cacheErr:  errors.New("redis down"),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pg.ExpectedCalls = nil
			redis.ExpectedCalls = nil
			minio.ExpectedCalls = nil

			// uploader
			minio.On("Upload", mock.Anything, int64(1), mock.Anything).
				Return(tt.uploadURL, tt.uploadErr)

			if tt.uploadErr == nil {
				pg.On("UpdatePhoto", mock.Anything, int64(1), tt.uploadURL).
					Return(tt.repoErr)

				if tt.repoErr == nil {
					redis.On("Invalidate", mock.Anything, int64(1)).
						Return(tt.cacheErr)
				}
			}

			url, err := uc.UploadPhoto(context.Background(), 1, bytes.NewReader([]byte("fake image")))
			if tt.expectErr && err == nil {
				t.Errorf("expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectErr && url != tt.uploadURL {
				t.Errorf("got url %s, want %s", url, tt.uploadURL)
			}

			pg.AssertExpectations(t)
			redis.AssertExpectations(t)
			minio.AssertExpectations(t)
		})
	}
}
