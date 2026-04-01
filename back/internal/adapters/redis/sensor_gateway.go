package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type sensorGateway struct {
	repo[domain.SensorGateway]
}

var _ SensorGateway = (*sensorGateway)(nil)

func NewSensorGateway(client *HealthyClient, defaultTTL time.Duration) *sensorGateway {
	return &sensorGateway{
		repo: repo[domain.SensorGateway]{
			HealthyClient: client,
			Key:           cache.SensorInstanceKey,
			DefaultTTL:    defaultTTL,
		},
	}
}

func (r *sensorGateway) CmdSetOne(ctx context.Context, entity *domain.SensorGateway) (*StatusCmd, error) {
	return r.repo.CmdSetOneByStrID(ctx, entity.ID.String(), entity)
}

func (r *sensorGateway) GetOneByID(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error) {
	return r.repo.GetOneByStrID(ctx, id.String())
}
