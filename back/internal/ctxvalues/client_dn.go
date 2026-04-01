package ctxvalues

import (
	"context"

	"github.com/brice-74/sensorflow/internal/types"
)

type clientDNKeyType struct{}

var clientDNKey = clientDNKeyType{}

func WithClientDN(ctx context.Context, dn *types.ClientDN) context.Context {
	return context.WithValue(ctx, clientDNKey, dn)
}

func GetClientDN(ctx context.Context) (*types.ClientDN, bool) {
	c, ok := ctx.Value(clientDNKey).(*types.ClientDN)
	return c, ok
}
