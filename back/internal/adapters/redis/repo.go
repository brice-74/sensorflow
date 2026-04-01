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
	*HealthyClient
	Key        cache.EntityKey
	KeyFunc    func(id string) string
	DefaultTTL time.Duration
}

var _ Repo[any] = (*repo[any])(nil)

func (r *repo[_]) Client() Client {
	return r.HealthyClient
}

func (r *repo[T]) key(id string) string {
	if r.KeyFunc != nil {
		return r.KeyFunc(id)
	}
	return cache.FormatKey(r.Key, id)
}

func (r *repo[T]) keys(ids ...string) []string {
	if r.KeyFunc != nil {
		result := make([]string, len(ids))
		for i, id := range ids {
			result[i] = r.KeyFunc(id)
		}
		return result
	}
	return cache.FormatKeys(r.Key, ids)
}

//
// --- Context helpers ---
//

func (r *repo[_]) UnaryCmdableFromCtx(ctx context.Context) redis.Cmdable {
	if conn, ok := GetConn(ctx); ok && conn != nil {
		return conn
	}
	return r.HealthyClient
}

func (r *repo[_]) PipelineFromCtx(ctx context.Context) (redis.Pipeliner, bool) {
	if pipe, ok := GetPipeline(ctx); ok && pipe != nil {
		return pipe, ok
	}
	return nil, false
}

func (r *repo[_]) CmdableFromCtx(ctx context.Context) redis.Cmdable {
	if pipe, ok := r.PipelineFromCtx(ctx); ok {
		return pipe
	}
	return r.UnaryCmdableFromCtx(ctx)
}

//
// --- Unexported helpers ---
//

func (r *repo[T]) cmdSetMany(ctx context.Context, cmdable redis.Cmdable, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd, error) {
	if len(entities) == 0 {
		return nil, nil, nil
	}

	pairs := make([]any, 0, len(entities)*2)
	keys := make([]string, 0, len(entities))

	for id, entity := range entities {
		data, err := gobutil.Encode(entity)
		if err != nil {
			return nil, nil, err
		}

		key := r.key(id)
		keys = append(keys, key)
		pairs = append(pairs, key, data)
	}

	statusCmd := &StatusCmd{
		Cmd:      cmdable.MSet(ctx, pairs...),
		markDown: r.MarkDown,
	}

	var boolCmds *MultiBoolCmd
	if ttl > 0 {
		boolCmds = &MultiBoolCmd{Cmds: make(map[string]*redis.BoolCmd, len(keys))}
		for _, key := range keys {
			boolCmds.Cmds[key] = cmdable.Expire(ctx, key, ttl)
		}
	}

	return statusCmd, boolCmds, nil
}

func (r *repo[T]) cmdSetIDsByParentID(ctx context.Context, cmdable redis.Cmdable, parentID string, ids []string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd) {
	if len(ids) == 0 {
		return nil, nil
	}

	key := cache.FormatHasManyKey(r.Key, parentID, relKey)
	intcmd := &IntCmd{
		Cmd:      cmdable.SAdd(ctx, key, ids),
		markDown: r.MarkDown,
	}

	var boolcmd *BoolCmd
	if ttl > 0 {
		boolcmd = &BoolCmd{
			Cmd:      cmdable.Expire(ctx, key, ttl),
			markDown: r.MarkDown,
		}
	}

	return intcmd, boolcmd
}

func (r *repo[T]) cmdSetOne(ctx context.Context, cmdable redis.Cmdable, id string, entity *T, ttl time.Duration) (*StatusCmd, error) {
	data, err := gobutil.Encode(entity)
	if err != nil {
		return nil, err
	}
	return &StatusCmd{
		Cmd:      cmdable.Set(ctx, r.key(id), data, ttl),
		markDown: r.MarkDown,
	}, nil
}

func (r *repo[T]) cmdListIDsByParentID(ctx context.Context, cmdable redis.Cmdable, id string, relKey cache.EntityKey) *StringSliceCmd {
	return &StringSliceCmd{
		Cmd:      cmdable.SMembers(ctx, cache.FormatHasManyKey(r.Key, id, relKey)),
		markDown: r.MarkDown,
	}
}

func (r *repo[T]) cmdGetOneByID(ctx context.Context, cmdable redis.Cmdable, id string) *StringCmdGob[T] {
	return &StringCmdGob[T]{Cmd: cmdable.Get(ctx, r.key(id)), markDown: r.MarkDown}
}

func (r *repo[T]) cmdGetByIDs(ctx context.Context, cmdable redis.Cmdable, ids []string) *SliceCmdGob[T] {
	return &SliceCmdGob[T]{Cmd: cmdable.MGet(ctx, r.keys(ids...)...), markDown: r.MarkDown}
}

//
// --- Commands ---
//

func (r *repo[T]) CmdSetStrIDsByParentID(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey) (*IntCmd, *BoolCmd) {
	return r.cmdSetIDsByParentID(ctx, r.Cmdable(ctx), parentID, ids, relKey, r.DefaultTTL)
}

func (r *repo[T]) CmdSetStrIDsByParentIDTTL(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd) {
	return r.cmdSetIDsByParentID(ctx, r.Cmdable(ctx), parentID, ids, relKey, ttl)
}

func (r *repo[T]) CmdSetManyByStrID(ctx context.Context, entities map[string]*T) (*StatusCmd, *MultiBoolCmd, error) {
	return r.cmdSetMany(ctx, r.Cmdable(ctx), entities, r.DefaultTTL)
}

func (r *repo[T]) CmdSetManyByStrIDTTL(ctx context.Context, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd, error) {
	return r.cmdSetMany(ctx, r.Cmdable(ctx), entities, ttl)
}

func (r *repo[T]) CmdSetOneByStrID(ctx context.Context, id string, entity *T) (*StatusCmd, error) {
	return r.cmdSetOne(ctx, r.Cmdable(ctx), id, entity, r.DefaultTTL)
}

func (r *repo[T]) CmdSetOneByStrIDTTL(ctx context.Context, id string, entity *T, ttl time.Duration) (*StatusCmd, error) {
	return r.cmdSetOne(ctx, r.Cmdable(ctx), id, entity, ttl)
}

func (r *repo[T]) CmdGetOneByStrID(ctx context.Context, id string) *StringCmdGob[T] {
	return r.cmdGetOneByID(ctx, r.Cmdable(ctx), id)
}

func (r *repo[T]) CmdGetManyByStrIDs(ctx context.Context, ids []string) *SliceCmdGob[T] {
	return r.cmdGetByIDs(ctx, r.Cmdable(ctx), ids)
}

func (r *repo[_]) CmdListStrIDsByParentID(ctx context.Context, id string, relKey cache.EntityKey) *StringSliceCmd {
	return r.cmdListIDsByParentID(ctx, r.Cmdable(ctx), id, relKey)
}

//
// --- Unary calls ---
//

func (r *repo[T]) SetStrIDsByParentID(ctx context.Context, parentID string, ids []string, relKey cache.EntityKey) error {
	insertCmd, ttlCmd := r.cmdSetIDsByParentID(ctx, r.Cmdable(ctx), parentID, ids, relKey, r.DefaultTTL)
	if _, err := insertCmd.Result(); err != nil {
		return err
	}
	if _, err := ttlCmd.Result(); err != nil {
		return err
	}
	return nil
}

func (r *repo[T]) SetManyByStrID(ctx context.Context, entities map[string]*T) error {
	insertCmd, ttlCmds, err := r.cmdSetMany(ctx, r.UnaryCmdableFromCtx(ctx), entities, r.DefaultTTL)
	if err != nil {
		return err
	}
	if _, err := insertCmd.Result(); err != nil {
		return err
	}
	if _, err := ttlCmds.Result(); err != nil {
		return err
	}
	return nil
}

func (r *repo[T]) SetOneByStrID(ctx context.Context, id string, entity *T) error {
	cmd, err := r.cmdSetOne(ctx, r.UnaryCmdableFromCtx(ctx), id, entity, r.DefaultTTL)
	if err != nil {
		return err
	}
	_, err = cmd.Result()
	return err
}

func (r *repo[T]) GetOneByStrID(ctx context.Context, id string) (*T, error) {
	return r.cmdGetOneByID(ctx, r.UnaryCmdableFromCtx(ctx), id).Result()
}

func (r *repo[T]) GetManyByStrIDs(ctx context.Context, ids []string) ([]*T, error) {
	return r.cmdGetByIDs(ctx, r.UnaryCmdableFromCtx(ctx), ids).Result()
}

func (r *repo[T]) GetManyByParentStrID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error) {
	cmdable := r.UnaryCmdableFromCtx(ctx)

	ids, err := r.cmdListIDsByParentID(ctx, cmdable, parentID, relKey).Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	return r.cmdGetByIDs(ctx, cmdable, ids).Result()
}
