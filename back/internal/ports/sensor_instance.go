package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstanceRepository interface {
	ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error)
	GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorInstance, error)
}
