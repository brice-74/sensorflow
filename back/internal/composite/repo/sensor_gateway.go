package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	dbRepo      *postgres.SensorGateway
	redisRepo   *redis.SensorGateway
	localCache  cache.Local[string, *domain.SensorGateway]
	invalidator cache.Invalidator[string]
}

var _ ports.SensorGatewayRepo = (*SensorGateway)(nil)

func NewSensorGateway() *SensorGateway {
	return &SensorGateway{}
}

func (r *SensorGateway) GetOneByID(ctx context.Context, id ulid.ULID) (*domain.SensorGateway, error) {
	key := cache.FormatKey(cache.SensorGatewayKey, id.String())

	if val, ok := r.localCache.Get(key); ok {
		return val, nil
	}

	if val, err := r.redisRepo.GetOneByID(ctx, id); err == nil {
		r.localCache.Set(key, val)
		return val, nil
	}

	val, err := r.dbRepo.GetOneByID(ctx, id)
	if err != nil {
		return nil, err
	}

	r.localCache.Set(key, val)
	return val, nil
}
