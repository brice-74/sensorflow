package ctxvalues

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type authSensorGatewayKeyType struct{}

var authSensorGatewayKey = authSensorGatewayKeyType{}

func WithAuthSensorGateway(ctx context.Context, c *SensorGatewayAuthContext) context.Context {
	return context.WithValue(ctx, authSensorGatewayKey, c)
}

func GetAuthSensorGateway(ctx context.Context) (*SensorGatewayAuthContext, bool) {
	c, ok := ctx.Value(authSensorGatewayKey).(*SensorGatewayAuthContext)
	return c, ok
}

type SensorGatewayAuthContext struct {
	GatewayID uuid.UUID
	TenantID  uuid.UUID
	// uuid -> sensor instance id
	// value -> sensor instance active plan
	Sensors map[uuid.UUID]domain.SensorPlanType
}
