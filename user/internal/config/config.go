package config

import (
	"log"
	"os"
)

type config struct {
	PostgresDSN      string
	RedisDSN         string
	MINIO_ENDPOINT   string
	MINIO_ACCESS_KEY string
	MINIO_SECRET_KEY string
	MINIO_BUCKET     string
	GRPC_PORT        string
}

var C config

func Load() {
	C = config{
		PostgresDSN:      getEnv("USER_POSTGRES_DSN", ""),
		RedisDSN:         getEnv("USER_REDIS_DSN", ""),
		MINIO_ENDPOINT:   getEnv("USER_MINIO_ENDPOINT", ""),
		MINIO_ACCESS_KEY: getEnv("USER_MINIO_ACCESS_KEY", ""),
		MINIO_SECRET_KEY: getEnv("USER_MINIO_SECRET_KEY", ""),
		MINIO_BUCKET:     getEnv("USER_MINIO_BUCKET", "users"),
		GRPC_PORT:        getEnv("USER_GRPC_PORT", ":50051"),
	}

	log.Println("✅ Config loaded")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
