package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	repo[ulid.ULID, *domain.SensorInstance[ulid.ULID]]
}

var _ ports.SensorInstanceRepository = (*SensorInstance)(nil)

func NewSensorInstance(client *HealthyClient) *SensorInstance {
	return &SensorInstance{
		repo: repo[domain.SensorInstance]{
			Rdb: client,
			Key: cache.SensorInstanceKey,
		},
	}
}

func (r *SensorInstance) CmdListIDsByGatewayID(ctx context.Context, gatewayID string) *StringSliceCmd {
	return r.CmdListIDsByParentID(ctx, gatewayID, cache.SensorGatewayKey)
}

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	return r.GetManyByParentID(ctx, gatewayID.String(), cache.SensorGatewayKey)
}
