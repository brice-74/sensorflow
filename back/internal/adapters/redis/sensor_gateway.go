package redis

import (
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type SensorGateway struct {
	repo[domain.SensorGateway]
}

var _ ports.SensorGatewayRepo = (*SensorGateway)(nil)

func NewSensorGateway(client *redis.Client) *SensorGateway {
	return &SensorGateway{
		repo: repo[domain.SensorGateway]{
			Client: client,
			Key:    cache.SensorInstanceKey,
		},
	}
}
