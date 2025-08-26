package repository

import (
	"app/user/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

//type CacheRepo interface {
//	SetProfile(user *domain.User) error
//	GetProfile(userID int64) (*domain.User, error)
//}

type RedisDB struct {
	client *redis.Client
}

func NewRedisDB(client *redis.Client) *RedisDB {
	return &RedisDB{client: client}
}

func (db *RedisDB) SetProfile(ctx context.Context, user *entity.User) error {
	b, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("user:%d", user.ID)
	return db.client.Set(ctx, key, b, 5*time.Minute).Err()
}

func (db *RedisDB) GetProfile(ctx context.Context, userID int64) (*entity.User, error) {
	key := fmt.Sprintf("user:%d", userID)
	b, err := db.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("not found")
		} else {
			return nil, err
		}
	}
	var user entity.User
	err = json.Unmarshal(b, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *RedisDB) Invalidate(ctx context.Context, userID int64) error {
	key := fmt.Sprintf("user:%d", userID)
	return db.client.Del(ctx, key).Err()
}
