package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstanceRepo interface {
	ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error)
}

type SensorGatewayRepo interface {
	GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error)
}

type SensorGatewayService interface {
	GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (*domain.SensorGateway, error)
}
