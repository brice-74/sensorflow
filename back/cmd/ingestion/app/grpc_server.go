package app

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// todo: compose GRPC services
type server struct {
}

func ServeGRPC(logger log.FiberLoggerInterface, cfg config.GRPC) error {
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

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
		// use futur middleware SensorAuth
		// grpc.StreamInterceptor(),
		// grpc.UnaryInterceptor(),
	)

	shutdownError := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// todo: here register GRPC services
		if err := s.Serve(lis); err != nil {
			shutdownError <- fmt.Errorf("server listen error: %w", err)
		}
	}()

	select {
	case sig := <-signalChan:
		logger.Info("shutting down gRPC server", log.Tags{"server_signal": sig.String()})

		done := make(chan struct{})
		go func() {
			s.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			logger.Info("gRPC server stopped gracefully", nil)
		case <-time.After(cfg.GracefulStopTimeout):
			logger.Info("gRPC server shutdown timed out, forcing stop", nil)
			s.Stop()
		}

		return nil
	case err := <-shutdownError:
		return err
	}
}
