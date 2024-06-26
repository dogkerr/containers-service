package db

import (
	"context"
	"dogker/lintang/container-service/config"
	"net/url"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(cfg *config.Config) *Postgres {
	dsn := url.URL{
		Scheme: cfg.Postgres.PGScheme,
		Host:   cfg.Postgres.PGURL, // "localhost:5432"
		User:   url.UserPassword(cfg.Postgres.Username, cfg.Postgres.Password),
		Path:   cfg.Postgres.PGDB,
	}

	q := dsn.Query()
	q.Add("sslmode", "disable")

	dsn.RawQuery = q.Encode()

	// // db, err := pgx.Connect(context.Background(), dsn.String())

	// db, err := sql.Open("pgx", dsn.String())
	// if err != nil {
	// 	hlog.Fatal("sql.Open", zap.Error(err))
	// }
	// db.SetMaxIdleConns(20)
	// db.SetMaxOpenConns(400) // awalnya 250
	// db.SetConnMaxIdleTime(5 * time.Minute)
	// db.SetConnMaxLifetime(60 * time.Minute)
	//  host := "postgres://" + cfg.Postgres.PGURL
	dbConfig, err := pgxpool.ParseConfig(dsn.String())
	dbConfig.MaxConns = 10
	dbConfig.MinConns = 2
	dbConfig.ConnConfig.ConnectTimeout = 40 * time.Second
	
	pool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		zap.L().Fatal("pgxpool connect", zap.Error(err))
	}

	if err := pool.Ping(context.Background()); err != nil {
		hlog.Fatal("db.PingContext", zap.Error(err))
	}
	return &Postgres{Pool: pool}
}

func (pg *Postgres) ClosePostgres(ctx context.Context) {
	zap.L().Info("closing postgres gracefully")
	pg.Pool.Close()
}
