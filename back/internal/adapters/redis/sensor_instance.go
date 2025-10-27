package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"github.com/redis/go-redis/v9"
)

type SensorInstance struct {
	repo[domain.SensorInstance]
}

var _ ports.SensorInstanceRepo = (*SensorInstance)(nil)

func NewSensorInstance(client *redis.Client) *SensorInstance {
	return &SensorInstance{
		repo: repo[domain.SensorInstance]{
			Client: client,
			Key:    cache.SensorInstanceKey,
		},
	}
}

func (r *SensorInstance) CmdListIDsByGatewayID(ctx context.Context, gatewayID string) *redis.StringSliceCmd {
	return r.CmdListIDsByParentID(ctx, gatewayID, cache.SensorGatewayKey)
}

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	return r.GetManyByParentID(ctx, gatewayID.String(), cache.SensorGatewayKey)
}
