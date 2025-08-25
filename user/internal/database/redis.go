package database

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
)

func ConnectRedis(addr string, password string, dbStr string, ctx context.Context) (*redis.Client, error) {
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверяем, что соединение с Redis работает
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("✅ Redis connected")
	return client, nil
}
