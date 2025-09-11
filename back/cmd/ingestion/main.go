package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/brice-74/sensorflow/cmd/ingestion/app"
	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/pkg/log"
)

func main() {
	cfg, err := app.ParseConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse config error:", err)
		os.Exit(1)
	}

	logger, flushSentry, err := app.PrepareLogger(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init logger:", err)
		os.Exit(1)
	}
	defer flushSentry()

	logger = logger.With(
		log.Tags{"api_identifier": cfg.Instance.Identifier},
	).(log.FiberLoggerInterface)

	pgclient := openPostgres(&cfg.Postgres, logger)
	sqlxDB := pgclient.Sqlx()

	repos := app.NewRepositories(sqlxDB)
	uows := app.NewUnitOfWorks(sqlxDB)
	svcs := app.NewServices(repos, uows)

	if err := app.ServeGRPC(cfg.GRPC, app.GRPCDeps{
		Logger:               logger,
		SensorGatewayService: svcs.SensorGateway,
	}); err != nil {
		logger.Error(err, log.Tags{"server": "closed"})
	}
}

func openPostgres(cfg *config.Postgres, logger log.LoggerInterface) postgres.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := postgres.NewClient(ctx, cfg)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	return client
}
