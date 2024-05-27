package webapi

import (
	"context"
	"dogker/lintang/container-service/pb"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type AuthClient struct {
	service pb.UsersServiceClient
}

func NewUserClient(cc *grpc.ClientConn) *AuthClient {
	service := pb.NewUsersServiceClient(cc)
	return &AuthClient{service: service}
}

func (m *AuthClient) GetUser(ctx context.Context, userID string) error {
	grpcCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.GetUserRequest{
		Id: userID,
	}

	_, err := m.service.GetUserById(grpcCtx, req)
	if err != nil {
		zap.L().Debug("m.service.GetUserById", zap.Error(err))
		return err
	}
	return nil
}
