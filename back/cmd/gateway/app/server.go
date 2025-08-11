package app

import (
	"bytes"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/http"
	"github.com/brice-74/sensorflow/pkg/log"
	"github.com/gofiber/fiber/v2"
)

type Server struct {
	*fiber.App
	logger log.FiberLoggerInterface
}

func NewFiberServer(logger log.FiberLoggerInterface) *Server {
	return &Server{
		logger: logger,
		App: fiber.New(
			fiber.Config{
				JSONDecoder: func(data []byte, v any) error {
					decoder := json.NewDecoder(bytes.NewReader(data))
					decoder.DisallowUnknownFields()
					return decoder.Decode(v)
				},
				ReadTimeout: time.Second * 5,
				BodyLimit:   5 * 1024 * 1024, // 5 Mo
				ErrorHandler: func(c *fiber.Ctx, err error) error {
					logger.WithFiberCtx(c).Error(err)
					return http.JSONInternalError(c, "fiber_unexpected_error", "Unexpected server error", nil)
				},
			},
		),
	}
}

func (srv *Server) Serve(port string) error {
	shutdownError := make(chan error)

	go func(srv *Server) {
		listenSignalOS := make(chan os.Signal, 1)
		signal.Notify(listenSignalOS, syscall.SIGINT, syscall.SIGTERM)
		signalOS := <-listenSignalOS

		srv.logger.Info("shutting down server", log.Tags{"signal": signalOS.String()})

		shutdownError <- srv.Shutdown()
	}(srv)

	if err := srv.Listen(":" + port); err != nil {
		return err
	}

	err := <-shutdownError
	if err != nil {
		return err
	}

	return nil
}
