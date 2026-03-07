package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Client interface {
	redis.Cmdable
	Cmdable(ctx context.Context) redis.Cmdable
	NewConn(ctx context.Context) (redis.Cmdable, context.Context)
	NewPipeline(ctx context.Context) (redis.Pipeliner, context.Context)
	HandleError(err error) error
	IsAlive() bool
}

type Repo[T any] interface {
	Client
	Client() Client
	// --- Context helpers ---
	CmdableFromCtx(ctx context.Context) redis.Cmdable
	PipelineFromCtx(ctx context.Context) (redis.Pipeliner, bool)
	UnaryCmdableFromCtx(ctx context.Context) redis.Cmdable
	// --- Commands ---
	CmdGetManyByStrIDs(ctx context.Context, ids []string) *SliceCmdGob[T]
	CmdGetOneByStrID(ctx context.Context, id string) *StringCmdGob[T]
	CmdListStrIDsByParentID(ctx context.Context, parentID string, relKey cache.EntityKey) *StringSliceCmd
	CmdSetStrIDsByParentID(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey) (*IntCmd, *BoolCmd)
	CmdSetStrIDsByParentIDTTL(ctx context.Context, ids []string, parentID string, relKey cache.EntityKey, ttl time.Duration) (*IntCmd, *BoolCmd)
	CmdSetManyByStrID(ctx context.Context, entities map[string]*T) (*StatusCmd, *MultiBoolCmd)
	CmdSetManyByStrIDTTL(ctx context.Context, entities map[string]*T, ttl time.Duration) (*StatusCmd, *MultiBoolCmd)
	CmdSetOneByStrID(ctx context.Context, id string, entity *T) *StatusCmd
	CmdSetOneByStrIDTTL(ctx context.Context, id string, entity *T, ttl time.Duration) *StatusCmd
	// --- Unary calls ---
	GetManyByStrIDs(ctx context.Context, ids []string) ([]*T, error)
	GetOneByStrID(ctx context.Context, id string) (*T, error)
	GetManyByParentStrID(ctx context.Context, parentID string, relKey cache.EntityKey) ([]*T, error)
	SetStrIDsByParentID(ctx context.Context, parentID string, ids []string, relKey cache.EntityKey) error
	SetManyByStrID(ctx context.Context, entities map[string]*T) error
	SetOneByStrID(ctx context.Context, id string, entity *T) error
}

type SensorPlanBinding interface {
	Repo[domain.SensorPlanBinding]
	ports.SensorPlanBindingRepository
}

type SensorGateway interface {
	Repo[domain.SensorGateway]
	ports.SensorGatewayRepository
	CmdSetOne(ctx context.Context, entity *domain.SensorGateway) *StatusCmd
}

type SensorInstance interface {
	Repo[domain.SensorInstance]
	ports.SensorInstanceRepository
	CmdSetMany(ctx context.Context, insts []*domain.SensorInstance) (*StatusCmd, *MultiBoolCmd)
	CmdListIDsByGatewayID(ctx context.Context, gtwID string) *StringSliceCmd
	SetIDsByGatewayID(ctx context.Context, gtwID uuid.UUID, insts []*domain.SensorInstance) error
}
