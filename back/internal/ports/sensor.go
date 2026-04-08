package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/types"
	"github.com/google/uuid"
)

type SensorGatewayRepository interface {
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error)
}

type SensorGatewayOrchestrator interface {
	GetOneWithInstances(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error)
}

type SensorInstanceRepository interface {
	ListByGatewayID(ctx context.Context, gatewayID uuid.UUID) ([]*domain.SensorInstance, error)
	GetOneByID(ctx context.Context, id uuid.UUID) (*domain.SensorInstance, error)
}

type SensorGatewayAuthRepository interface {
	GetAuthPlanRowsByGatewayID(ctx context.Context, gatewayID uuid.UUID) ([]*types.SensorGatewayAuthPlanRow, error)
}

type SensorGatewayAuthOrchestrator interface {
	GetOneByID(ctx context.Context, ID uuid.UUID) (*ctxvalues.SensorGatewayAuthContext, error)
}
