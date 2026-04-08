package middleware

import (
	"context"

	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/internal/types"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/google/uuid"
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
	log            log.Logger
	svcGatewayAuth ports.SensorGatewayAuthOrchestrator
}

func NewMTLSClientAuth(logger log.Logger, svcGatewayAuth ports.SensorGatewayAuthOrchestrator) *MTLSClientAuth {
	return &MTLSClientAuth{
		log:            logger,
		svcGatewayAuth: svcGatewayAuth,
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

	gw, err := m.loadSensorGatewayCtx(ctx, client.CN)
	if err != nil {
		return nil, err
	}
	ctx = ctxvalues.WithAuthSensorGateway(ctx, gw)

	return ctx, nil
}

func (m *MTLSClientAuth) loadSensorGatewayCtx(ctx context.Context, cn string) (*ctxvalues.SensorGatewayAuthContext, error) {
	id, err := uuid.Parse(cn)
	if err != nil {
		m.log.Error(errors.Wrap(err, "invalid uuid"), log.Contexts{"dn": {"cn": cn}})
		return nil, status.Error(codes.Unauthenticated, "invalid sensor gateway ID")
	}

	sg, err := m.svcGatewayAuth.GetOneByID(ctx, id)
	if err != nil {
		var e *errors.Error
		if errors.As(err, &e) {
			switch e.Code {
			case errors.CodeTimeout,
				errors.CodeCanceled,
				errors.CodeInternal:
				break
			default:
				return nil, status.Error(codes.Unauthenticated, "invalid sensor gateway ID")
			}
		}
		m.log.Error(errors.Wrap(err, "failed to get sensor gateway context"), log.Contexts{"sensor_gateway": {"id": id}})
		return nil, status.Error(codes.Internal, "something wen't wrong")
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
