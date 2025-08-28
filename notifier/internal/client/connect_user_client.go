package client

import (
	"context"
	"time"

	userpb "app/user/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectUserClient(ctx context.Context, addr string) (userpb.UserServiceClient, *grpc.ClientConn, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, nil, err
	}

	client := userpb.NewUserServiceClient(conn)
	return client, conn, nil
}
