package postgres

import (
	"context"
	"fmt"
	"net/url"

	"github.com/brice-74/sensorflow/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Client interface {
	Pgx() *pgxpool.Pool
	Sqlx() *sqlx.DB
	Close()
}

type client struct {
	pgxPool *pgxpool.Pool
	sqlxDB  *sqlx.DB
}

// creates a client that allows you to take advantage of the features of Pgx and Sqlx while having a single connection pool
func NewClient(ctx context.Context, cfg *config.Postgres) (Client, error) {
	dsn := buildDSN(cfg)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = int32(cfg.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.MaxIdleConns)
	poolCfg.MaxConnIdleTime = cfg.ConnMaxIdleTime
	poolCfg.MaxConnLifetime = cfg.ConnMaxLifetime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	if cfg.InitialPingTimeout > 0 {
		ctxTimeout, cancel := context.WithTimeout(ctx, cfg.InitialPingTimeout)
		defer cancel()
		if err := pool.Ping(ctxTimeout); err != nil {
			return nil, err
		}
	} else {
		if err := pool.Ping(ctx); err != nil {
			return nil, err
		}
	}

	sqlxdb := sqlx.NewDb(stdlib.OpenDBFromPool(pool), "pgx")

	return &client{
		sqlxDB:  sqlxdb,
		pgxPool: pool,
	}, nil
}

// Since the pool is shared, simply close it from pgx.
func (c *client) Close()             { c.pgxPool.Close() }
func (c *client) Pgx() *pgxpool.Pool { return c.pgxPool }
func (c *client) Sqlx() *sqlx.DB     { return c.sqlxDB }

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
