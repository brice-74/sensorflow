package app

import (
	"time"

	ristrettoadapter "github.com/brice-74/sensorflow/internal/adapters/ristretto"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/orchestrators"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/dgraph-io/ristretto/v2"
)

type Orchestrators struct {
	SensorGateway ports.SensorGatewayOrchestrator
}

func NewOrchestrators(
	redisRepos *RedisRepositories,
	pgRepos *PostgresRepositories,
	localCacheHub *ristretto.Cache[string, any],
	logger log.Logger,
	asyncPool ports.AsyncSubmitter,
) *Orchestrators {
	SensorGatewayLocalCache := ristrettoadapter.NewLocalCache[*domain.SensorGateway](localCacheHub, 1*time.Minute)

	orcts := Orchestrators{
		SensorGateway: orchestrators.NewSensorGateway(
			pgRepos.SensorGateway,
			pgRepos.SensorInstance,
			redisRepos.SensorGateway,
			redisRepos.SensorInstance,
			SensorGatewayLocalCache,
			logger,
			asyncPool,
		),
	}
	return &orcts
}
