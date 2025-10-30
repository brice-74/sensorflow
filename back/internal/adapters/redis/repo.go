package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/redis/go-redis/v9"
)

// All Redis helper methods prefixed with 'Cmd' are intended for use in any context,
// including pipelines, and therefore require manual execution.
// Other helper methods are meant for single, unary calls.
type repo[T any] struct {
	Rdb        *HealthyClient
	Key        cache.EntityKey
	DefaultTTL time.Duration
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

func (r *repo[T]) cmdSetMany(ctx context.Context, cmdable redis.Cmdable, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd) {
	if len(entities) == 0 {
		return nil, nil
	}

	pairs := make([]any, 0, len(entities)*2)
	keys := make([]string, 0, len(entities))

	for id, entity := range entities {
		data, err := gobutil.Encode(entity)
		if err != nil {
			return nil, nil
		}

		key := cache.FormatKey(r.Key, id)
		keys = append(keys, key)
		pairs = append(pairs, key, data)
	}

	statusCmd := &StatusCmd{
		Cmd:      cmdable.MSet(ctx, pairs...),
		markDown: r.Rdb.markDown,
	}

	var boolCmds *MultiBoolCmd
	if ttl > 0 {
		boolCmds = &MultiBoolCmd{Cmds: make(map[string]*redis.BoolCmd, len(keys))}
		for _, key := range keys {
			boolCmds.Cmds[key] = cmdable.Expire(ctx, key, ttl)
		}
	}

	return statusCmd, boolCmds
}

func (r *repo[T]) cmdSetIDsByParentID(ctx context.Context, cmdable redis.Cmdable, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd) {
	if len(ids) == 0 {
		return nil, nil
	}

	key := cache.FormatHasManyKey(r.Key, parentID, relKey)
	intcmd := &IntCmd{
		Cmd:      cmdable.SAdd(ctx, key, ids),
		markDown: r.Rdb.markDown,
	}

	var boolcmd *BoolCmd
	if ttl > 0 {
		boolcmd = &BoolCmd{
			Cmd:      cmdable.Expire(ctx, key, ttl),
			markDown: r.Rdb.markDown,
		}
	}

	return intcmd, boolcmd
}

func (r *repo[T]) cmdSetOne(ctx context.Context, cmdable redis.Cmdable, id string, entity *T, ttl time.Duration) *StatusCmd {
	return &StatusCmd{
		Cmd:      cmdable.Set(ctx, cache.FormatKey(r.Key, id), entity, ttl),
		markDown: r.Rdb.markDown,
	}
}

func (r *repo[T]) cmdListIDsByParentID(ctx context.Context, cmdable redis.Cmdable, id string, relKey cache.EntityKey) *StringSliceCmd {
	return &StringSliceCmd{
		Cmd:      cmdable.SMembers(ctx, cache.FormatHasManyKey(r.Key, id, relKey)),
		markDown: r.Rdb.markDown,
	}
}

func (r *repo[T]) cmdGetOneByID(ctx context.Context, cmdable redis.Cmdable, id string) *StringCmdGob[T] {
	return &StringCmdGob[T]{Cmd: cmdable.Get(ctx, cache.FormatKey(r.Key, id)), markDown: r.Rdb.markDown}
}

func (r *repo[T]) cmdGetByIDs(ctx context.Context, cmdable redis.Cmdable, ids []string) *SliceCmdGob[T] {
	return &SliceCmdGob[T]{Cmd: cmdable.MGet(ctx, cache.FormatKeys(cache.SensorInstanceKey, ids)...), markDown: r.Rdb.markDown}
}

//
// --- Commands ---
//

func (r *repo[T]) CmdSetIDsByParentID(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey) (*IntCmd, *BoolCmd) {
	return r.cmdSetIDsByParentID(ctx, r.Cmdable(ctx), ids, parentID, relKey, r.DefaultTTL)
}

func (r *repo[T]) CmdSetIDsByParentIDTTL(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd) {
	return r.cmdSetIDsByParentID(ctx, r.Cmdable(ctx), ids, parentID, relKey, ttl)
}

func (r *repo[T]) CmdSetMany(ctx context.Context, entities map[string]*T) (*StatusCmd, *MultiBoolCmd) {
	return r.cmdSetMany(ctx, r.Cmdable(ctx), entities, r.DefaultTTL)
}

func (r *repo[T]) CmdSetManyTTL(ctx context.Context, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd) {
	return r.cmdSetMany(ctx, r.Cmdable(ctx), entities, ttl)
}

func (r *repo[T]) CmdSetOne(ctx context.Context, id string, entity *T) *StatusCmd {
	return r.cmdSetOne(ctx, r.Cmdable(ctx), id, entity, r.DefaultTTL)
}

func (r *repo[T]) CmdSetOneTTL(ctx context.Context, id string, entity *T, ttl time.Duration) *StatusCmd {
	return r.cmdSetOne(ctx, r.Cmdable(ctx), id, entity, ttl)
}

func (r *repo[T]) CmdGetOneByID(ctx context.Context, id string) *StringCmdGob[T] {
	return r.cmdGetOneByID(ctx, r.Cmdable(ctx), id)
}

func (r *repo[T]) CmdGetByIDs(ctx context.Context, ids []string) *SliceCmdGob[T] {
	return r.cmdGetByIDs(ctx, r.Cmdable(ctx), ids)
}

func (r *repo[_]) CmdListIDsByParentID(ctx context.Context, id string, relKey cache.EntityKey) *StringSliceCmd {
	return r.cmdListIDsByParentID(ctx, r.Cmdable(ctx), id, relKey)
}

//
// --- Unary calls ---
//

func (r *repo[T]) SetMany(ctx context.Context, entities map[string]*T) error {
	insertCmd, ttlsCmds := r.cmdSetMany(ctx, r.UnaryCmdable(ctx), entities, r.DefaultTTL)
	if _, err := insertCmd.Result(); err != nil {
		return err
	}
	if _, err := ttlsCmds.Result(); err != nil {
		return err
	}
	return nil
}

func (r *repo[T]) SetOne(ctx context.Context, id string, entity *T) error {
	_, err := r.cmdSetOne(ctx, r.UnaryCmdable(ctx), id, entity, r.DefaultTTL).Result()
	return err
}

func (r *repo[T]) GetOneByID(ctx context.Context, id string) (*T, error) {
	return r.cmdGetOneByID(ctx, r.UnaryCmdable(ctx), id).Result()
}

func (r *repo[T]) GetByIDs(ctx context.Context, ids []string) ([]*T, error) {
	return r.cmdGetByIDs(ctx, r.UnaryCmdable(ctx), ids).Result()
}

func (r *repo[T]) GetManyByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error) {
	cmdable := r.UnaryCmdable(ctx)

	ids, err := r.cmdListIDsByParentID(ctx, cmdable, parentID, relKey).Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	return r.cmdGetByIDs(ctx, cmdable, ids).Result()
}
