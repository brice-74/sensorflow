package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"github.com/redis/go-redis/v9"
)

// All Redis helper methods prefixed with 'Cmd' are intended for use in any context,
// including pipelines, and therefore require manual execution.
// Other helper methods are meant for single, unary calls.
type repo[T any] struct {
	Rdb *HealthyClient
	Key cache.EntityKey
}

//
// --- Context helpers ---
//

func (r *repo[_]) UnaryCmdable(ctx context.Context) redis.Cmdable {
	if conn, ok := GetConn(ctx); ok && conn != nil {
		return conn
	}
	return r.Rdb.Client
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
// --- Unexported helpers ---
//

func (r *repo[T]) cmdGetOneByID(ctx context.Context, cmdable redis.Cmdable, id string) *StringCmd[T] {
	return &StringCmd[T]{Cmd: cmdable.Get(ctx, cache.FormatKey(r.Key, id))}
}

func (*repo[T]) cmdGetByIDs(ctx context.Context, cmdable redis.Cmdable, ids []string) *SliceCmd[T] {
	return &SliceCmd[T]{Cmd: cmdable.MGet(ctx, cache.FormatKeys(cache.SensorInstanceKey, ids)...)}
}

func (r *repo[T]) cmdListIDsByParentID(ctx context.Context, cmdable redis.Cmdable, id string, relKey cache.EntityKey) *redis.StringSliceCmd {
	return cmdable.SMembers(ctx, cache.FormatHasManyKey(r.Key, id, relKey))
}

//
// --- Commands ---
//

func (r *repo[T]) CmdGetOneByID(ctx context.Context, id string) *StringCmd[T] {
	return r.cmdGetOneByID(ctx, r.Cmdable(ctx), id)
}

func (r *repo[T]) CmdGetByIDs(ctx context.Context, ids []string) *SliceCmd[T] {
	return r.cmdGetByIDs(ctx, r.Cmdable(ctx), ids)
}

func (r *repo[_]) CmdListIDsByParentID(ctx context.Context, id string, relKey cache.EntityKey) *redis.StringSliceCmd {
	return r.cmdListIDsByParentID(ctx, r.Cmdable(ctx), id, relKey)
}

//
// --- Unary calls ---
//

func (r *repo[T]) GetOneByID(ctx context.Context, id ulid.ULID) (*T, error) {
	res, err := r.cmdGetOneByID(ctx, r.UnaryCmdable(ctx), id.String()).Result()
	return res, r.Rdb.HandleError(err)
}

func (r *repo[T]) GetByIDs(ctx context.Context, ids []ulid.ULID) ([]*T, error) {
	res, err := r.cmdGetByIDs(ctx, r.UnaryCmdable(ctx), ulid.ToStrings(ids)).Result()
	return res, r.Rdb.HandleError(err)
}

func (r *repo[T]) GetManyByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error) {
	cmdable := r.UnaryCmdable(ctx)

	ids, err := r.cmdListIDsByParentID(ctx, cmdable, parentID, relKey).Result()
	if err != nil || len(ids) == 0 {
		return nil, r.Rdb.HandleError(err)
	}

	res, err := r.cmdGetByIDs(ctx, cmdable, ids).Result()
	return res, r.Rdb.HandleError(err)
}
