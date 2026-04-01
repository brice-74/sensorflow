package orchestrators

import (
	"context"
	"time"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/worker"
	"github.com/google/uuid"
)

type SensorGatewayAuthOrchestrator struct {
	dbRepo     ports.SensorGatewayAuthRepository
	redisRepo  redisadapter.SensorGatewayAuth
	localCache cache.Local[*ctxvalues.SensorGatewayContext]
	logger     log.Logger
	asyncPool  ports.AsyncSubmitter[ports.Task]
}

func (o *SensorGatewayAuthOrchestrator) GetOne(ctx context.Context, gatewayID uuid.UUID) (*ctxvalues.SensorGatewayContext, error) {
	cacheKey := cache.FormatSensorGatewayAuthKey(gatewayID.String())

	if gw, ok := o.localCache.Get(cacheKey); ok {
		return gw, nil
	}

	if redisCli := o.redisRepo.Client(); redisCli.IsAlive() {
		gw, err := o.redisRepo.GetOneByID(ctx, gatewayID)
		err = o.redisRepo.HandleError(err)
		if err != nil {
			o.logger.Warn(errors.WrapErr(err))
		} else if gw != nil {
			o.localCache.Set(cacheKey, gw)
			return gw, nil
		}
	}

	gw, err := o.dbRepo.GetOneByID(ctx, gatewayID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}
	if gw == nil {
		return nil, errors.WrapMsg("Nil SensorGatewayContext")
	}

	o.localCache.Set(cacheKey, gw)

	if err := o.asyncPool.Submit(worker.VoidTask(func() {
		ctxBg, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := o.redisRepo.SetOne(ctxBg, gw); err != nil {
			o.logger.Warn(errors.WrapErr(err))
		}
	})); err != nil {
		o.logger.Warn(errors.WrapErr(err))
	}

	return gw, nil
}
