package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type SensorInstanceRepository interface {
	ListByGatewayID(ctx context.Context, gatewayID uuid.UUID) ([]*domain.SensorInstance, error)
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.SensorInstance, error)
}
