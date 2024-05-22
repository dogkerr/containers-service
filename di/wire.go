// go:build wireinject
//go:build wireinject
// +build wireinject

package di

import (
	"dogker/lintang/container-service/biz/dal/db"
	"dogker/lintang/container-service/biz/dal/messagebroker"
	"dogker/lintang/container-service/biz/router"
	"dogker/lintang/container-service/biz/service"
	"dogker/lintang/container-service/biz/webapi"
	"dogker/lintang/container-service/config"

	"github.com/google/wire"
	"google.golang.org/grpc"
)

var ProviderSet wire.ProviderSet = wire.NewSet(
	service.NewContainerService,
	webapi.CreateNewDockerEngineAPI,
	db.NewContainerRepo,
	webapi.CreateDkronAPI,
	webapi.NewMonitorClient,
	webapi.NewMinioAPI,

	wire.Bind(new(router.ContainerService), new(*service.ContainerService)),
	wire.Bind(new(service.ContainerRepository), new(*db.ContainerRepository)),
	wire.Bind(new(service.DockerEngineAPI), new(*webapi.DockerEngineAPI)),
	wire.Bind(new(service.DkronAPI), new(*webapi.DkronAPI)),
	wire.Bind(new(service.MonitorClient), new(*webapi.MonitorClient)),
	wire.Bind(new(service.MinioAPI), new(*webapi.MinioAPI)),
)

func InitContainerService(pg *db.Postgres, rmq *messagebroker.RabbitMQ, cfg *config.Config,
	cc *grpc.ClientConn) *service.ContainerService {
	wire.Build(
		ProviderSet,
	)
	return nil
}

// kitex
var ProviderSetGRPC wire.ProviderSet = wire.NewSet(
	service.NewContainerGRPCService,
	webapi.CreateNewDockerEngineAPI,
	db.NewContainerRepo,
	webapi.CreateDkronAPI,
	webapi.NewMonitorClient,
	webapi.NewMinioAPI,
	wire.Bind(new(service.ContainerRepository), new(*db.ContainerRepository)),
	wire.Bind(new(service.DockerEngineAPI), new(*webapi.DockerEngineAPI)),
	wire.Bind(new(service.DkronAPI), new(*webapi.DkronAPI)),
	wire.Bind(new(service.MonitorClient), new(*webapi.MonitorClient)),
	wire.Bind(new(service.MinioAPI), new(*webapi.MinioAPI)),
)

func InitContainerGRPCService(pg *db.Postgres, rmq *messagebroker.RabbitMQ, cfg *config.Config,
	cc *grpc.ClientConn) *service.ContainerGRPCServiceImpl {
	wire.Build(
		ProviderSetGRPC,
	)
	return nil
}
