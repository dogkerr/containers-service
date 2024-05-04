package db

import (
	"context"
	"database/sql"
	"dogker/lintang/container-service/config"
	"net/url"

	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.uber.org/zap"
)

type Postgres struct {
	Pool *sql.DB
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

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		hlog.Fatal("sql.Open", zap.Error(err))
	}

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(250)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(60 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
		hlog.Fatal("db.PingContext", zap.Error(err))
	}
	return &Postgres{db}
}

func ClosePostgres(pg *sql.DB) error {
	err := pg.Close()
	return err
}
