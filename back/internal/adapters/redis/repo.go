package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"github.com/redis/go-redis/v9"
)

// All Redis helpers methods that begin with Cmd are designed to be used in all contexts, including pipelines.
// The others are designed for unary calls.
type repo[T any] struct {
	Client *redis.Client
	Key    cache.EntityKey
}

//
// --- Context helpers ---
//

func (r *repo[_]) UnaryCmdable(ctx context.Context) redis.Cmdable {
	if conn, ok := GetConn(ctx); ok && conn != nil {
		return conn
	}
	return r.Client
}

func (r *repo[_]) Pipeline(ctx context.Context) (redis.Pipeliner, bool) {
	if pipe, ok := GetPipeline(ctx); ok && pipe != nil {
		return pipe, ok
	}
	return nil, false
}

func (r *repo[_]) Cmdable(ctx context.Context) redis.Cmdable {
	if pipe, ok := r.Pipeline(ctx); ok {
		return pipe
	}
	return r.UnaryCmdable(ctx)
}

//
// --- Internal Redis helpers ---
//

func (r *repo[T]) cmdGetOneByID(ctx context.Context, cmdable redis.Cmdable, id string) *StringCmd[T] {
	return &StringCmd[T]{Cmd: cmdable.Get(ctx, cache.FormatKey(r.Key, id))}
}

func (r *repo[T]) CmdGetOneByID(ctx context.Context, id string) *StringCmd[T] {
	return r.cmdGetOneByID(ctx, r.Cmdable(ctx), id)
}

func (r *repo[T]) GetOneByID(ctx context.Context, id ulid.ULID) (*T, error) {
	return r.cmdGetOneByID(ctx, r.UnaryCmdable(ctx), id.String()).Result()
}

func (*repo[T]) cmdGetByIDs(ctx context.Context, cmdable redis.Cmdable, ids []string) *SliceCmd[T] {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = cache.FormatKey(cache.SensorInstanceKey, id)
	}
	return &SliceCmd[T]{Cmd: cmdable.MGet(ctx, keys...)}
}

func (r *repo[T]) CmdGetByIDs(ctx context.Context, ids []string) *SliceCmd[T] {
	return r.cmdGetByIDs(ctx, r.Cmdable(ctx), ids)
}

func (r *repo[T]) GetByIDs(ctx context.Context, ids []ulid.ULID) ([]*T, error) {
	return r.cmdGetByIDs(ctx, r.UnaryCmdable(ctx), ulid.ToStrings(ids)).Result()
}

func (r *repo[T]) cmdListIDsByParentID(ctx context.Context, cmdable redis.Cmdable, id string, relKey cache.EntityKey) *redis.StringSliceCmd {
	return cmdable.SMembers(ctx, cache.FormatHasManyKey(
		r.Key, id, relKey,
	))
}

func (r *repo[T]) CmdListIDsByParentID(ctx context.Context, id string, relKey cache.EntityKey) *redis.StringSliceCmd {
	return r.Cmdable(ctx).SMembers(ctx, cache.FormatHasManyKey(
		r.Key, id, relKey,
	))
}

func (r *repo[T]) GetManyByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error) {
	idsCmd := r.cmdListIDsByParentID(ctx, r.UnaryCmdable(ctx), parentID, relKey)
	ids, err := idsCmd.Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	valuesCmd := r.cmdGetByIDs(ctx, r.UnaryCmdable(ctx), ids)
	return valuesCmd.Result()
}
