package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/redis/go-redis/v9"
)

const SensorInstanceKey redisadapter.EntityKey = "sensor_instance"

type SensorInstance struct {
	*redisadapter.Repo
}

var _ ports.SensorInstanceRepo = (*SensorInstance)(nil)

func NewSensorInstance(repo *redisadapter.Repo) *SensorInstance {
	return &SensorInstance{
		Repo: repo,
	}
}

func (r *SensorInstance) CmdListIDsByGatewayID(ctx context.Context, gatewayID ulid.ULID) *redis.StringSliceCmd {
	cmdable := r.Cmdable(ctx)
	relationKey := redisadapter.FormatHasManyKey(
		SensorGatewayKey, gatewayID.String(), SensorInstanceKey,
	)
	return cmdable.SMembers(ctx, relationKey)
}

func (r *SensorInstance) CmdGetMany(ctx context.Context, ids []string) *redis.SliceCmd {
	cmdable := r.Cmdable(ctx)

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = redisadapter.FormatEntityKey(SensorInstanceKey, id)
	}

	return cmdable.MGet(ctx, keys...)
}

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	cmdable := r.Cmdable(ctx)

	relationKey := redisadapter.FormatHasManyKey(
		SensorGatewayKey, gatewayID.String(), SensorInstanceKey,
	)

	ids, err := cmdable.SMembers(ctx, relationKey).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = redisadapter.FormatEntityKey(SensorInstanceKey, id)
	}

	values, err := cmdable.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.SensorInstance, 0, len(values))
	for _, val := range values {
		if val == nil {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		inst, err := gobutil.Decode[domain.SensorInstance]([]byte(str))
		if err != nil {
			continue
		}
		result = append(result, inst)
	}

	return result, nil
}
