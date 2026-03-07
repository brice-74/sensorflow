package orchestrators

import (
	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
)

type SensorPlanBinding struct {
	dbRepo     ports.SensorPlanBindingRepository
	redisRepo  redisadapter.SensorPlanBinding
	localCache cache.Local[*domain.SensorPlanBinding]
	logger     log.Logger
}
