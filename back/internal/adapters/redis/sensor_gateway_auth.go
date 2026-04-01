package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/google/uuid"
)

type sensorGatewayAuth struct {
	repo[ctxvalues.SensorGatewayContext]
}

var _ SensorGatewayAuth = (*sensorGatewayAuth)(nil)

func NewSensorGatewayAuth(client *HealthyClient, defaultTTL time.Duration) *sensorGatewayAuth {
	return &sensorGatewayAuth{
		repo: repo[ctxvalues.SensorGatewayContext]{
			HealthyClient: client,
			KeyFunc:       cache.FormatSensorGatewayAuthKey,
			DefaultTTL:    defaultTTL,
		},
	}
}

func (r *sensorGatewayAuth) GetOneByID(ctx context.Context, id uuid.UUID) (*ctxvalues.SensorGatewayContext, error) {
	return r.repo.GetOneByStrID(ctx, id.String())
}

func (r *sensorGatewayAuth) SetOne(ctx context.Context, gw *ctxvalues.SensorGatewayContext) error {
	return r.repo.SetOneByStrID(ctx, gw.GatewayID.String(), gw)
}
