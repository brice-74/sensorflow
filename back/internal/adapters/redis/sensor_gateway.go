package redis

import (
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ports"
)

type SensorGateway struct {
	repo[domain.SensorGateway]
}

var _ ports.SensorGatewayRepository = (*SensorGateway)(nil)

func NewSensorGateway(client *HealthyClient) *SensorGateway {
	return &SensorGateway{
		repo: repo[domain.SensorGateway]{
			Rdb: client,
			Key: cache.SensorInstanceKey,
		},
	}
}
