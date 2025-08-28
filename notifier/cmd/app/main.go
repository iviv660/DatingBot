package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/notifier/internal"
	"app/notifier/internal/client"
	"app/notifier/internal/config"
	"app/notifier/internal/tg"
)

func main() {
	config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userCli, _, err := client.ConnectUserClient(ctx, config.C.UserGRPCAddr)
	if err != nil {
		log.Fatalf("user grpc: %v", err)
	}

	matchCli, _, err := client.ConnectMatchClient(ctx, config.C.MatchGRPCAddr)
	if err != nil {
		log.Fatalf("match grpc: %v", err)
	}

	userAdapter := client.NewUserClientAdapter(userCli)
	matchAdapter := client.NewMatchClientAdapter(matchCli)

	core := internal.NewCore(userAdapter, matchAdapter)

	bot, err := tg.NewBot(config.C.TelegramToken)
	if err != nil {
		log.Fatalf("telebot init: %v", err)
	}

	h := tg.NewHandler(bot, core)
	h.Register()

	go bot.Start()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	bot.Stop()
	time.Sleep(200 * time.Millisecond)
}
