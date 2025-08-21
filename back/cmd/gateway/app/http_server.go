package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brice-74/sensorflow/internal/adapters/http"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/gofiber/fiber/v2"
)

func ServeHTTP(logger log.FiberLoggerInterface, cfg config.HTTP) error {
	fiberApp := fiber.New(
		fiber.Config{
			JSONDecoder: func(data []byte, v any) error {
				decoder := json.NewDecoder(bytes.NewReader(data))
				decoder.DisallowUnknownFields()
				return decoder.Decode(v)
			},
			IdleTimeout:      cfg.IdleTimeout,
			Concurrency:      cfg.ConcurrencyLimit,
			DisableKeepalive: !cfg.Keepalive,
			WriteTimeout:     cfg.WriteTimeout,
			ReadTimeout:      cfg.ReadTimeout,
			BodyLimit:        int(cfg.BodyLimit),
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				logger.WithFiberCtx(c).Error(err)
				return http.JSONInternalError(c, "fiber_unexpected_error", "Unexpected server error", nil)
			},
		},
	)

	shutdownError := make(chan error, 1)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := fiberApp.Listen(":" + cfg.Port); err != nil {
			shutdownError <- fmt.Errorf("server listen error: %w", err)
		}
	}()

	select {
	case sig := <-signalChan:
		logger.Info("shutting down HTTP server", log.Tags{"server_signal": sig.String()})

		if err := fiberApp.ShutdownWithTimeout(cfg.GracefulStopTimeout); err != nil {
			return fmt.Errorf("shutdown error: %w", err)
		}
		return nil
	case err := <-shutdownError:
		return err
	}
}
