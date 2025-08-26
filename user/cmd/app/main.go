package main

import (
	"app/user/internal/config"
	"app/user/internal/database"
	"app/user/internal/handler"
	"app/user/internal/repository"
	"app/user/internal/usecase"
	userpb "app/user/proto"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	config.Load()
	postgresCon, err := database.ConnectPostgres(config.C.PostgresDSN, context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer postgresCon.Close()
	postgres := repository.NewPostgresDB(postgresCon)

	redisCon, err := database.ConnectRedis(context.Background(), config.C.RedisDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer redisCon.Close()
	redis := repository.NewRedisDB(redisCon)

	minioCon, err := database.ConnectMinio(
		context.Background(),
		config.C.MINIO_ENDPOINT,
		config.C.MINIO_ACCESS_KEY,
		config.C.MINIO_SECRET_KEY)
	if err != nil {
		log.Fatal(err)
	}
	minio := repository.NewMinio(minioCon, config.C.MINIO_BUCKET)

	uc := usecase.New(postgres, redis, minio)
	h := handler.NewHandler(uc)

	lis, err := net.Listen("tcp", config.C.GRPC_PORT)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	userpb.RegisterUserServiceServer(grpcServer, h)
	reflection.Register(grpcServer)
	log.Println("✅ gRPC UserService running on", config.C.GRPC_PORT)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
