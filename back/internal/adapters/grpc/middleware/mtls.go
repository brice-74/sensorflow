package middleware

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/types"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ClientDNHeaderKey = "x-client-dn"

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

type MTLSClientAuth struct {
	log              log.LoggerInterface
	svcSensorGateway ports.SensorGatewayService
}

func NewMTLSClientAuth(logger log.LoggerInterface, svcSensorGateway ports.SensorGatewayService) *MTLSClientAuth {
	return &MTLSClientAuth{
		log:              logger,
		svcSensorGateway: svcSensorGateway,
	}
}

func (m *MTLSClientAuth) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		ctx := ss.Context()

		ctx, err = m.handleContext(ctx)
		if err != nil {
			return err
		}

		return handler(srv, &wrappedStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func (m *MTLSClientAuth) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		var err error

		ctx, err = m.handleContext(ctx)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (m *MTLSClientAuth) handleContext(ctx context.Context) (context.Context, error) {
	client, err := m.extractClientDN(ctx)
	if err != nil {
		return nil, err
	}
	ctx = ctxvalues.WithClientDN(ctx, client)

	gateway, err := m.loadSensorGateway(ctx, client.CN)
	if err != nil {
		return nil, err
	}
	ctx = ctxvalues.WithSensorGateway(ctx, gateway)

	return ctx, nil
}

func (m *MTLSClientAuth) loadSensorGateway(ctx context.Context, cn string) (*domain.SensorGateway, error) {
	id, err := ulid.Parse(cn)
	if err != nil {
		m.log.Error(errors.Wrap(err, "invalid ulid"), log.Contexts{"dn": {"cn": cn}})
		return nil, status.Errorf(codes.Unauthenticated, "invalid sensor gateway ID")
	}

	sg, err := m.svcSensorGateway.GetOneWithInstances(ctx, id)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "sensor gateway not found")
		}
		m.log.Error(errors.Wrap(err, "failed to get sensor gateway"), log.Contexts{"sensor_gateway": {"id": id}})
		return nil, status.Errorf(codes.Internal, "failed to retrieve sensor gateway")
	}
	return sg, nil
}

func (*MTLSClientAuth) extractClientDN(ctx context.Context) (*types.ClientDN, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(ClientDNHeaderKey)
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing client DN header")
	}

	client, err := types.ParseClientDN(values[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid client DN: %v", err)
	}

	return client, nil
}
