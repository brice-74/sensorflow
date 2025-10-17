package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/brice-74/sensorflow/pkg/ulid"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/redis/go-redis/v9"
)

type SensorInstance struct {
	*redisadapter.Repo
}

var _ ports.SensorInstanceRepo = (*SensorInstance)(nil)

func NewSensorInstance(repo *redisadapter.Repo) *SensorInstance {
	return &SensorInstance{
		Repo: repo,
	}
}

func (*SensorInstance) cmdListIDsByGatewayID(ctx context.Context, cmdable redis.Cmdable, gatewayID ulid.ULID) *redis.StringSliceCmd {
	relationKey := cache.FormatHasManyKey(
		cache.SensorGatewayKey, gatewayID.String(), cache.SensorInstanceKey,
	)
	return cmdable.SMembers(ctx, relationKey)
}

func (r *SensorInstance) CmdListIDsByGatewayID(ctx context.Context, gatewayID ulid.ULID) *redis.StringSliceCmd {
	return r.cmdListIDsByGatewayID(ctx, r.Cmdable(ctx), gatewayID)
}

func (*SensorInstance) cmdGetByIDs(ctx context.Context, cmdable redis.Cmdable, ids []string) *redis.SliceCmd {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = cache.FormatEntityKey(cache.SensorInstanceKey, id)
	}
	return cmdable.MGet(ctx, keys...)
}

func (r *SensorInstance) CmdGetByIDs(ctx context.Context, ids []string) *redis.SliceCmd {
	return r.cmdGetByIDs(ctx, r.Cmdable(ctx), ids)
}

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	idsCmd := r.cmdListIDsByGatewayID(ctx, r.UnaryCmdable(ctx), gatewayID)
	ids, err := idsCmd.Result()
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	valuesCmd := r.cmdGetByIDs(ctx, r.UnaryCmdable(ctx), ids)
	values, err := valuesCmd.Result()
	if err != nil {
		return nil, err
	}

	return gobutil.DecodeManyAnyStr[domain.SensorInstance](values)
}
