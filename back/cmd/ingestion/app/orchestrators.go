package app

import (
	"time"

	ristrettoadapter "github.com/brice-74/sensorflow/internal/adapters/ristretto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/orchestrators"
	"github.com/brice-74/sensorflow/internal/orchestrators/ingestor"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/dgraph-io/ristretto/v2"
)

type Orchestrators struct {
	SensorGatewayAuth ports.SensorGatewayAuthOrchestrator

	IngestAccelGyroStd ports.Ingestor[*domain.AccelGyroMeasurement, ingestor.IngestStatus]
	/*
		 	IngestAccelGyroIndustrial ports.Ingestor[*domain.AccelGyroMeasurement, ingestor.IngestStatus]
			IngestAccelGyroRealtime   ports.Ingestor[*domain.AccelGyroMeasurement, ingestor.IngestStatus]
	*/
}

func NewOrchestrators(
	redisRepos *RedisRepositories,
	pgRepos *PostgresRepositories,
	clickhouseRepos *ClickhouseRepositories,
	localCacheHub *ristretto.Cache[string, any],
	logger log.Logger,
	asyncPool ports.AsyncSubmitter[ports.Task],
) (*Orchestrators, error) {
	sensorGatewayAuthLocalCache := ristrettoadapter.NewLocalCache[*ctxvalues.SensorGatewayAuthContext](localCacheHub, 1*time.Minute)

	ingestAccelGyroStd, err := ingestor.New(ingestor.Config[*domain.AccelGyroMeasurement]{
		Name:   "ingest-accel-gyro-std",
		Logger: logger,
		Levels: []ingestor.IngestionLevel[*domain.AccelGyroMeasurement]{
			ingestor.NewClickhouseLevel(clickhouseRepos.AccelGyroStd, logger, 100, 5*time.Minute),
		},
	})
	if err != nil {
		return nil, err
	}

	ingestAccelGyroStd.Start()

	orcts := Orchestrators{
		SensorGatewayAuth: orchestrators.NewSensorGatewayAuth(
			pgRepos.SensorGatewayAuth,
			redisRepos.SensorGatewayAuth,
			sensorGatewayAuthLocalCache,
			logger,
			asyncPool,
		),
		IngestAccelGyroStd: ingestAccelGyroStd,
	}

	return &orcts, nil
}
