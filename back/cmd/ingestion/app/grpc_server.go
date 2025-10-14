package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/grpc/middleware"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// todo: compose GRPC services
type server struct {
}

type GRPCDeps struct {
	Logger               log.Fiber
	SensorGatewayService ports.SensorGatewayService
}

func ServeGRPC(cfg config.GRPC, deps GRPCDeps) error {
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	tlsAuth := middleware.NewMTLSClientAuth(deps.Logger, deps.SensorGatewayService)

	kaParams := keepalive.ServerParameters{
		MaxConnectionIdle:     cfg.IdleTimeout,
		MaxConnectionAge:      cfg.MaxConnectionAge,
		MaxConnectionAgeGrace: cfg.MaxConnectionAgeGrace,
		Time:                  cfg.KeepaliveTime,
		Timeout:               cfg.KeepaliveTimeout,
	}

	kaPolicy := keepalive.EnforcementPolicy{
		MinTime:             cfg.MinTimeBetweenPings,
		PermitWithoutStream: cfg.AllowPingWithoutActiveRPCs,
	}

	s := grpc.NewServer(
		grpc.KeepaliveParams(kaParams),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
		grpc.MaxConcurrentStreams(cfg.MaxConcurrentStreams),
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgSize),
		grpc.ChainStreamInterceptor(
			tlsAuth.StreamInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			tlsAuth.UnaryInterceptor(),
		),
	)

	// todo: here register GRPC services

	shutdownError := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.Serve(lis); err != nil {
			shutdownError <- fmt.Errorf("server listen error: %w", err)
		}
	}()

	select {
	case sig := <-signalChan:
		deps.Logger.Info("shutting down gRPC server", log.Tags{"server_signal": sig.String()})

		done := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			deps.Logger.Info("gRPC server stopped gracefully")
		case <-time.After(cfg.GracefulStopTimeout):
			deps.Logger.Info("gRPC server shutdown timed out, forcing stop")
			s.Stop()
		}

		return nil
	case err := <-shutdownError:
		return err
	}
}
