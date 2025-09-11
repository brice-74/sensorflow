package ctxvalues

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/types"
)

type clientDNKey struct{}

func WithClientDN(ctx context.Context, dn *types.ClientDN) context.Context {
	return context.WithValue(ctx, clientDNKey{}, dn)
}

func GetClientDN(ctx context.Context) (*types.ClientDN, bool) {
	c, ok := ctx.Value(clientDNKey{}).(*types.ClientDN)
	return c, ok
}

type sensorGatewayKey struct{}

func WithSensorGateway(ctx context.Context, sensorGateway *domain.SensorGateway) context.Context {
	return context.WithValue(ctx, sensorGatewayKey{}, sensorGateway)
}

func GetSensorGateway(ctx context.Context) (*domain.SensorGateway, bool) {
	c, ok := ctx.Value(sensorGatewayKey{}).(*domain.SensorGateway)
	return c, ok
}
