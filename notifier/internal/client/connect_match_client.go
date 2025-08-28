package client

import (
	"context"
	"time"

	matchpb "app/match/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectMatchClient(ctx context.Context, addr string) (matchpb.MatchServiceClient, *grpc.ClientConn, error) {
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

	client := matchpb.NewMatchServiceClient(conn)
	return client, conn, nil
}
