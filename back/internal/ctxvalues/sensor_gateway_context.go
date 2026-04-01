package ctxvalues

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type authSensorGatewayKeyType struct{}

var authSensorGatewayKey = authSensorGatewayKeyType{}

func WithAuthSensorGateway(ctx context.Context, sensorGateway *domain.SensorGateway) context.Context {
	c := SensorGatewayContext{
		GatewayID: sensorGateway.ID,
		Sensors:   make(map[uuid.UUID]*SensorContext, len(sensorGateway.SensorInstances)),
	}

	for _, v := range sensorGateway.SensorInstances {
		c.Sensors[v.ID] = &SensorContext{
			Status:     v.Status,
			Firmware:   v.Firmware,
			ActivePlan: v.ActiveSensorPlanBinding,
		}
	}

	return context.WithValue(ctx, authSensorGatewayKey, &c)
}

func GetAuthSensorGateway(ctx context.Context) (*SensorGatewayContext, bool) {
	c, ok := ctx.Value(authSensorGatewayKey).(*SensorGatewayContext)
	return c, ok
}

type SensorGatewayContext struct {
	GatewayID uuid.UUID
	Sensors   map[uuid.UUID]*SensorContext
}

type SensorContext struct {
	Status     domain.SensorStatus
	Firmware   *string
	ActivePlan *domain.SensorPlanBinding
}
