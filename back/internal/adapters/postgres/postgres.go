package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/brice-74/sensorflow/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
)

type Client struct {
	Pool *pgxpool.Pool
	DB   *sqlx.DB
}

func NewClient(ctx context.Context, cfg *config.Postgres) (*Client, error) {
	pool, err := OpenPGXPool(ctx, cfg)
	if err != nil {
		return nil, err
	}

	db, err := OpenSQLX(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		Pool: pool,
		DB:   db,
	}, nil
}

func OpenSQL(ctx context.Context, cfg *config.Postgres) (*sql.DB, error) {
	db, err := sql.Open("postgres", buildDSN(cfg))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if timeout := cfg.InitialPingTimeout; timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

func OpenSQLX(ctx context.Context, cfg *config.Postgres) (*sqlx.DB, error) {
	db, err := OpenSQL(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, "postgres"), nil
}

func OpenPGXPool(ctx context.Context, cfg *config.Postgres) (*pgxpool.Pool, error) {
	poolcfg, err := pgxpool.ParseConfig(buildDSN(cfg))
	if err != nil {
		return nil, err
	}

	poolcfg.MaxConns = int32(cfg.MaxOpenConns)
	poolcfg.MinConns = int32(cfg.MaxIdleConns)
	poolcfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolcfg.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolcfg)
	if err != nil {
		return nil, err
	}

	if timeout := cfg.InitialPingTimeout; timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func buildDSN(cfg *config.Postgres) string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.Database,
	}

	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()

	return u.String()
}
