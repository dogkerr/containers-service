// go:build wireinject
//go:build wireinject
// +build wireinject

package di

import (
	"dogker/lintang/container-service/biz/dal/db"
	"dogker/lintang/container-service/biz/dal/messagebroker"
	"dogker/lintang/container-service/biz/router"
	"dogker/lintang/container-service/biz/service"
	"dogker/lintang/container-service/config"

	"github.com/google/wire"
)

var ProviderSet wire.ProviderSet = wire.NewSet(
	service.NewContainerService,
	db.NewContainerRepo,

	wire.Bind(new(router.ContainerService), new(*service.ContainerService)),
	wire.Bind(new(service.ContainerRepository), new(*db.ContainerRepository)),
)

func InitContainerService(pg *db.Postgres, rmq *messagebroker.RabbitMQ, cfg *config.Config) *service.ContainerService {
	wire.Build(
		ProviderSet,
	)
	return nil
}
