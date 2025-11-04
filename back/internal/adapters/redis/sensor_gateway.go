package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type sensorGateway struct {
	repo[domain.SensorGateway]
}

var _ ports.SensorGatewayRepository = (*sensorGateway)(nil)

func NewSensorGateway(client *HealthyClient) *sensorGateway {
	return &sensorGateway{
		repo: repo[domain.SensorGateway]{
			Rdb: client,
			Key: cache.SensorInstanceKey,
		},
	}
}

func (r *sensorGateway) CmdSetOne(ctx context.Context, entity *domain.SensorGateway) *StatusCmd {
	return r.repo.CmdSetOne(ctx, entity.ID.String(), entity)
}

func (r *sensorGateway) GetOneByID(ctx context.Context, id ulid.ULID) (*domain.SensorGateway, error) {
	return r.repo.GetOneByID(ctx, id.String())
}
