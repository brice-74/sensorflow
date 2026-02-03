package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type SensorGatewayRepository interface {
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error)
}

type SensorGatewayOrchestrator interface {
	GetOneWithInstances(ctx context.Context, ID uuid.UUID) (*domain.SensorGateway, error)
}
