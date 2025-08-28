package config

import (
	"os"
)

type config struct {
	TelegramToken string
	UserGRPCAddr  string
	MatchGRPCAddr string
}

var C config

func Load() {
	C = config{
		TelegramToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		UserGRPCAddr:  getEnv("USER_CLIENT", "user_service:50051"),
		MatchGRPCAddr: getEnv("MATCH_CLIENT", "match_service:50052"),
	}

}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
