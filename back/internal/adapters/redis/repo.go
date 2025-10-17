package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Repo struct {
	Client *redis.Client
}

func (r *Repo) UnaryCmdable(ctx context.Context) redis.Cmdable {
	if conn, ok := GetConn(ctx); ok && conn != nil {
		return conn
	}
	return r.Client
}

func (r *Repo) Pipeline(ctx context.Context) (redis.Pipeliner, bool) {
	if pipe, ok := GetPipeline(ctx); ok && pipe != nil {
		return pipe, ok
	}
	return nil, false
}

func (r *Repo) Cmdable(ctx context.Context) redis.Cmdable {
	if pipe, ok := r.Pipeline(ctx); ok {
		return pipe
	}
	return r.UnaryCmdable(ctx)
}
