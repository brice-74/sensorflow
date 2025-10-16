package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type connCtxKeyType struct{}

var connCtxKey = connCtxKeyType{}

func WithConn(ctx context.Context, conn *redis.Conn) context.Context {
	return context.WithValue(ctx, connCtxKey, conn)
}

func GetConn(ctx context.Context) (*redis.Conn, bool) {
	conn, ok := ctx.Value(connCtxKey).(*redis.Conn)
	return conn, ok
}

type pipeCtxKeyType struct{}

var pipeCtxKey = pipeCtxKeyType{}

func WithPipeline(ctx context.Context, pipe redis.Pipeliner) context.Context {
	return context.WithValue(ctx, pipeCtxKey, pipe)
}

func GetPipeline(ctx context.Context) (redis.Pipeliner, bool) {
	pipe, ok := ctx.Value(pipeCtxKey).(redis.Pipeliner)
	return pipe, ok
}
