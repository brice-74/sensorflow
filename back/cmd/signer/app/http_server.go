package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/brice-74/sensorflow/internal/log"
)

type HTTPDeps struct {
	Logger  log.Logger
	Service *Service
}

func ServeHTTP(cfg Signer, deps HTTPDeps) error {
	handlers := NewHandlers(cfg, deps.Service, deps.Logger)
	mux := http.NewServeMux()
	registerRoutes(mux, handlers)

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	shutdownError := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			shutdownError <- fmt.Errorf("server listen error: %w", err)
		}
	}()

	select {
	case sig := <-signalChan:
		deps.Logger.Info("shutting down HTTP server", log.Tags{"server_signal": sig.String()})

		ctx, cancel := context.WithTimeout(context.Background(), cfg.GracefulStopTimeout)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-shutdownError:
		return err
	}
}
