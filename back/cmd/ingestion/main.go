package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/brice-74/sensorflow/cmd/ingestion/app"
	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/config"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/worker"
	"github.com/dgraph-io/ristretto/v2"
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
	).(log.Fiber)

	pgclient := openPostgres(&cfg.Postgres, logger)
	defer pgclient.Close()

	sqlxDB := pgclient.Sqlx()

	redisClient := redisadapter.NewHealthyClientFromCfg(&cfg.Redis)
	defer redisClient.Close()

	ristrettoCache := openLocalCache(logger)

	clickhouseClient := clickhouse.NewHealthyClient(&cfg.Clickhouse)
	defer clickhouseClient.Client.Close()

	asyncTasksPool := orchestrtorsAsyncTasks(cfg.OrchestrtorsAsyncTasksConfig, logger)

	clickhouseRepos := app.NewClickhouseRepositories(clickhouseClient)
	pgRepos := app.NewPostgresRepositories(sqlxDB)
	redisRepos := app.NewRedisRepositories(redisClient)
	orchestrators, err := app.NewOrchestrators(redisRepos, pgRepos, clickhouseRepos, ristrettoCache, logger, asyncTasksPool)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	if err := app.ServeGRPC(cfg.GRPC, app.GRPCDeps{
		Logger:                        logger,
		SensorGatewayAuthOrchestrator: orchestrators.SensorGatewayAuth,
	}); err != nil {
		logger.Error(err, log.Tags{"server": "closed"})
	}
}

func openPostgres(cfg *config.Postgres, logger log.Logger) postgres.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := postgres.NewClient(ctx, cfg)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	return client
}

func openLocalCache(logger log.Logger) *ristretto.Cache[string, any] {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters:        10_000,
		MaxCost:            1_000,
		BufferItems:        64,
		Metrics:            false,
		IgnoreInternalCost: true,
	})
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}
	return ristrettoCache
}

func orchestrtorsAsyncTasks(cfg app.OrchestrtorsAsyncTasksConfig, logger log.Logger) *worker.BoundedPool[ports.Task] {
	return worker.NewBoundedPool(
		worker.WithMinWorkers[ports.Task](cfg.WorkerPool.MinWorkers),
		worker.WithMaxWorkers[ports.Task](cfg.WorkerPool.MaxWorkers),
		worker.WithIdleTimeout[ports.Task](cfg.WorkerPool.IdleTimeout),
		worker.WithQueueSize[ports.Task](cfg.WorkerPool.QueueSize),
		worker.WithPanicHandler[ports.Task](log.PanicHandler(logger)),
	)
}
