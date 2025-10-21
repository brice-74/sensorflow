package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type GenericRepo[T any] interface {
	GetOneByID(ctx context.Context, id ulid.ULID) (*T, error)
}

type compositeRepo[T any, RR redis.Repo[T], DR GenericRepo[T]] struct {
	DBRepo      DR
	RedisRepo   RR
	LocalCache  cache.Local[string, *T]
	Invalidator cache.Invalidator[string]
}

type LayeredFallbackRead[T any, RR redis.Repo[T], DR GenericRepo[T]] struct {
	*compositeRepo[T, RR, DR]
}

func (r *LayeredFallbackRead[T, _, _]) GetOneByID(ctx context.Context, id ulid.ULID) (*T, error) {
	key := id.String()
	if val, ok := r.LocalCache.Get(key); ok {
		return val, nil
	}

	val, errRedis := r.RedisRepo.GetOneByID(ctx, id)
	if val != nil {
		r.LocalCache.Set(key, val)
		return val, errors.WrapErr(errRedis)
	}

	val, errDB := r.DBRepo.GetOneByID(ctx, id)
	if errDB != nil {
		return nil, errors.JoinWrap(errRedis, errDB)
	}

	r.LocalCache.Set(key, val)
	return val, errors.WrapErr(errRedis)
}
