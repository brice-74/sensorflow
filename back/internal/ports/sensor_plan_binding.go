package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type SensorPlanBindingRepository interface {
	GetActiveByInstanceID(ctx context.Context, instanceID uuid.UUID) (*domain.SensorPlanBinding, error)
}
