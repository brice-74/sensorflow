package orchestrators

import (
	"context"
	"time"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ctxvalues"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/internal/types"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/worker"
	"github.com/google/uuid"
)

type SensorGatewayAuth struct {
	dbRepo     ports.SensorGatewayAuthRepository
	redisRepo  redisadapter.SensorGatewayAuth
	localCache cache.Local[*ctxvalues.SensorGatewayAuthContext]
	logger     log.Logger
	asyncPool  ports.AsyncSubmitter[ports.Task]
}

var _ ports.SensorGatewayAuthOrchestrator = (*SensorGatewayAuth)(nil)

func NewSensorGatewayAuth(
	dbRepo ports.SensorGatewayAuthRepository,
	redisRepo redisadapter.SensorGatewayAuth,
	localCache cache.Local[*ctxvalues.SensorGatewayAuthContext],
	logger log.Logger,
	asyncPool ports.AsyncSubmitter[ports.Task],
) *SensorGatewayAuth {
	return &SensorGatewayAuth{
		dbRepo:     dbRepo,
		redisRepo:  redisRepo,
		localCache: localCache,
		logger:     logger,
		asyncPool:  asyncPool,
	}
}

func (o *SensorGatewayAuth) GetOneByID(ctx context.Context, gatewayID uuid.UUID) (*ctxvalues.SensorGatewayAuthContext, error) {
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

	refs, err := o.dbRepo.GetAuthPlanRowsByGatewayID(ctx, gatewayID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	gw := buildSensorGatewayAuthContext(gatewayID, refs)
	if gw == nil {
		return nil, errors.WrapMsg("Nil SensorGatewayAuthContext")
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

func buildSensorGatewayAuthContext(gatewayID uuid.UUID, sensors []*types.SensorGatewayAuthPlanRow) *ctxvalues.SensorGatewayAuthContext {
	if len(sensors) == 0 {
		return nil
	}

	gwCtx := &ctxvalues.SensorGatewayAuthContext{
		GatewayID: gatewayID,
		TenantID:  sensors[0].TenantID,
		Sensors:   make(map[uuid.UUID]domain.SensorPlanType, len(sensors)),
	}
	for _, sensor := range sensors {
		gwCtx.Sensors[sensor.SensorInstanceID] = sensor.ActivePlan
	}
	return gwCtx
}
