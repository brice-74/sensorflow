package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type sensorPlanBinding struct {
	repo[domain.SensorPlanBinding]
}

var _ SensorPlanBinding = (*sensorPlanBinding)(nil)

func NewSensorPlanBinding(client *HealthyClient, defaultTTL time.Duration) *sensorPlanBinding {
	return &sensorPlanBinding{
		repo: repo[domain.SensorPlanBinding]{
			HealthyClient: client,
			Key:           cache.SensorPlanBindingKey,
			DefaultTTL:    defaultTTL,
		},
	}
}

func (s *sensorPlanBinding) GetActiveByInstanceID(ctx context.Context, instanceID uuid.UUID) (*domain.SensorPlanBinding, error) {
	key := cache.FormatActivePlanKey(instanceID.String())
	return s.GetOneByStrID(ctx, key)
}

func (s *sensorPlanBinding) ListActiveByInstanceIDs(ctx context.Context, instanceIDs []uuid.UUID) ([]*domain.SensorPlanBinding, error) {
	keys := make([]string, len(instanceIDs))
	for i, id := range instanceIDs {
		keys[i] = cache.FormatActivePlanKey(id.String())
	}
	return s.GetManyByStrIDs(ctx, keys)
}

func (s *sensorPlanBinding) SetOne(ctx context.Context, spb *domain.SensorPlanBinding) error {
	key := cache.FormatActivePlanKey(spb.SensorInstanceID.String())
	return s.SetOneByStrID(ctx, key, spb)
}
