package config

import (
	"log"
	"os"
)

type config struct {
	PostgresDSN string
	GRPC_PORT   string
	USER_CLIENT string
}

var C config

func Load() {
	C = config{
		PostgresDSN: getEnv("MATCH_POSTGRES_DSN", ""),
		GRPC_PORT:   getEnv("MATCH_GRPC_PORT", ":50052"),
		USER_CLIENT: getEnv("MATCH_USER_CLIENT", "user_service:50051"),
	}

	log.Println("✅ Config loaded")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
