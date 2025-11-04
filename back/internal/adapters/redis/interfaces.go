package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"github.com/redis/go-redis/v9"
)

type RedisCmdable interface {
	Cmdable(ctx context.Context) redis.Cmdable
	Pipeline(ctx context.Context) (redis.Pipeliner, context.Context)
	HandleError(err error) error
	IsHealthy() bool
}

type Repo[T any] interface {
	// --- Redis Commandes PURES (retourne des types Cmd concrets) ---
	CmdGetManyByStrIDs(ctx context.Context, ids []string) *SliceCmdGob[T]
	CmdGetOneByStrID(ctx context.Context, id string) *StringCmdGob[T]

	CmdListStrIDsByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) *StringSliceCmd
	CmdSetStrIDsByParentID(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey) (*IntCmd, *BoolCmd)
	CmdSetStrIDsByParentIDTTL(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd)

	CmdSetMany(ctx context.Context, entities map[string]*T) (*StatusCmd, *MultiBoolCmd)
	CmdSetManyTTL(ctx context.Context, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd)

	CmdSetOneByStrID(ctx context.Context, id string, entity *T) *StatusCmd
	CmdSetOneByStrIDTTL(ctx context.Context, id string, entity *T, ttl time.Duration) *StatusCmd

	// --- Accès bas niveau Redis ---
	Cmdable(ctx context.Context) redis.Cmdable
	Pipeline(ctx context.Context) (redis.Pipeliner, bool)
	UnaryCmdable(ctx context.Context)

	// --- Méthodes fonctionnelles (plus haut niveau) ---
	GetManyByStrIDs(ctx context.Context, ids []string) ([]*T, error)
	GetOneByStrID(ctx context.Context, id string) (*T, error)
	GetManyByParentStrID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error)

	SetMany(ctx context.Context, entities map[string]*T) error
	SetOneByStrID(ctx context.Context, id string, entity *T) error
	SetStrIDsByParentID(ctx context.Context, parentID string, ids []string, relKey cache.EntityKey) error
}

type SensorGateway interface {
	Repo[domain.SensorGateway]
	CmdSetOne(ctx context.Context, entity *domain.SensorGateway) *StatusCmd
	GetOneByID(ctx context.Context, id ulid.ULID) (*domain.SensorGateway, error)
}
