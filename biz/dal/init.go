package dal

import (
	"dogker/lintang/container-service/biz/dal/db"
	"dogker/lintang/container-service/biz/dal/messagebroker"
	"dogker/lintang/container-service/config"
)

func InitPg(cfg *config.Config) *db.Postgres {
	pg := db.NewPostgres(cfg)
	return pg
}
func InitRmq(cfg *config.Config) *messagebroker.RabbitMQ {
	rmq := messagebroker.NewRabbitMQ(cfg)

	return rmq
}
