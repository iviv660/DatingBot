package main

import (
	"context"
	"log"
	"net"

	"app/match/internal/client"
	"app/match/internal/config"
	"app/match/internal/database"
	"app/match/internal/handler"
	"app/match/internal/repository"
	"app/match/internal/usecase"
	matchpb "app/match/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	config.Load()

	ctx := context.Background()

	pgConn, err := database.ConnectPostgres(ctx, config.C.PostgresDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer pgConn.Close()
	matchRepo := repository.NewPostgresDB(pgConn)

	userGRPC, userConn, err := client.ConnectUserClient(ctx, config.C.USER_CLIENT)
	if err != nil {
		log.Fatal(err)
	}
	defer userConn.Close()
	userClient := client.NewUserClientAdapter(userGRPC)

	uc := usecase.NewUseCase(matchRepo, userClient)
	h := handler.NewHandler(uc)

	lis, err := net.Listen("tcp", config.C.GRPC_PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	matchpb.RegisterMatchServiceServer(s, h)

	reflection.Register(s)

	log.Println("✅ gRPC MatchService running on", config.C.GRPC_PORT)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
