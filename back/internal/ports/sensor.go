package ports

import (
	"context"

	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/google/uuid"
)

type SensorGatewayAuthOrchestrator = SensorGatewayAuthRepository

type SensorGatewayAuthRepository interface {
	GetOneByID(ctx context.Context, ID uuid.UUID) (*ctxvalues.SensorGatewayContext, error)
}
