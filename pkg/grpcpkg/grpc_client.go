package grpcpkg

import (
	"dogker/lintang/container-service/config"
	"dogker/lintang/container-service/pb"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MonitorGRPCClient struct {
	MonitorClient pb.MonitorServiceClient
}

func InitMonitorServiceClient(cfg *config.Config) *MonitorGRPCClient {
	cc, err := grpc.NewClient(cfg.GRPC.MonitorURL+"?wait=30s", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		zap.L().Fatal("Ggagl konek ke monitor service pake GRPC", zap.Error(err))

	}
	res := &MonitorGRPCClient{
		pb.NewMonitorServiceClient(cc),
	}
	return res
	// pb.NewMonitorServiceClient(cc)
}
