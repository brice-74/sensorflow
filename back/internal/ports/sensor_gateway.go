package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGatewayRepository interface {
	GetOneByID(ctx context.Context, id ulid.ULID) (*domain.SensorGateway, error)
}

type SensorGatewayRedisRepository interface {
	SensorGatewayRepository
}
