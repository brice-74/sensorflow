package middleware

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ClientDNHeaderKey = "x-client-id"

type ClientDNContextKey struct{}

func GetClientDN(ctx context.Context) (*ClientDN, bool) {
	c, ok := ctx.Value(ClientDNContextKey{}).(*ClientDN)
	return c, ok
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

type SensorAuth struct {
	// todo: use futur sensor gateway authentication service
}

func (*SensorAuth) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		client, err := extractClientDN(ctx)
		if err != nil {
			return err
		}

		ctx = context.WithValue(ctx, ClientDNContextKey{}, client)
		return handler(srv, &wrappedStream{
			ServerStream: ss,
			ctx:          ctx,
		})
	}
}

func (*SensorAuth) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		client, err := extractClientDN(ctx)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, ClientDNContextKey{}, client)
		return handler(ctx, req)
	}
}

func extractClientDN(ctx context.Context) (*ClientDN, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get(ClientDNHeaderKey)
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing client DN header")
	}

	client, err := parseClientDN(values[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid client DN: %v", err)
	}

	return client, nil
}

type ClientDN struct {
	CN string
	OU string
	O  string
	L  string
	ST string
	C  string
}

func parseClientDN(dn string) (*ClientDN, error) {
	var res = new(ClientDN)
	start := 0
	l := len(dn)

	for i := 0; i <= l; i++ {
		if i == l || dn[i] == ',' {
			if i <= start+2 {
				return nil, fmt.Errorf("invalid DN segment: %q", dn[start:i])
			}

			part := dn[start:i]
			start = i + 1

			var key, val string
			for j := 0; j < len(part); j++ {
				if part[j] == '=' {
					if j == 0 || j == len(part)-1 {
						return nil, fmt.Errorf("invalid key=value pair: %q", part)
					}
					key = part[:j]
					val = part[j+1:]
					break
				}
			}

			if key == "" || val == "" {
				return nil, fmt.Errorf("invalid key=value pair: %q", part)
			}

			// validate -> ASCII >=32, != ',' ou '=')
			for k := 0; k < len(val); k++ {
				if val[k] < 32 || val[k] == ',' || val[k] == '=' {
					return nil, fmt.Errorf("invalid DN value for %s: %q", key, val)
				}
			}

			switch key {
			case "CN":
				res.CN = val
			case "OU":
				res.OU = val
			case "O":
				res.O = val
			case "L":
				res.L = val
			case "ST":
				res.ST = val
			case "C":
				res.C = val
			default:
				return nil, fmt.Errorf("invalid DN key: %q", key)
			}
		}
	}

	return res, nil
}
