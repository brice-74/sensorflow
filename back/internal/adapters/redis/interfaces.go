package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/redis/go-redis/v9"
)

type RedisCmdable interface {
	Cmdable(ctx context.Context) redis.Cmdable
	Pipeline(ctx context.Context) (redis.Pipeliner, context.Context)
	HandleError(err error) error
	IsHealthy() bool
}

type Repo[T any] interface {
	CmdGetByIDs(ctx context.Context, ids []string) *SliceCmdGob[T]
	CmdGetOneByID(ctx context.Context, id string) *StringCmdGob[T]
	CmdListIDsByParentID(ctx context.Context, id string, relKey cache.EntityKey) *StringSliceCmd
	CmdSetIDsByParentID(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey) (*IntCmd, *BoolCmd)
	CmdSetIDsByParentIDTTL(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd)
	CmdSetMany(ctx context.Context, entities map[string]*T) (*StatusCmd, *MultiBoolCmd)
	CmdSetManyTTL(ctx context.Context, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd)
	CmdSetOne(ctx context.Context, id string, entity *T) *StatusCmd
	CmdSetOneTTL(ctx context.Context, id string, entity *T, ttl time.Duration) *StatusCmd
	Cmdable(ctx context.Context) redis.Cmdable
	GetByIDs(ctx context.Context, ids []string) ([]*T, error)
	GetManyByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error)
	GetOneByID(ctx context.Context, id string) (*T, error)
	Pipeline(ctx context.Context) (redis.Pipeliner, bool)
	SetIDsByParentID(ctx context.Context, parentID string, ids []string, relKey cache.EntityKey) error
	SetMany(ctx context.Context, entities map[string]*T) error
	SetOne(ctx context.Context, id string, entity *T) error
	UnaryCmdable(ctx context.Context)
}
