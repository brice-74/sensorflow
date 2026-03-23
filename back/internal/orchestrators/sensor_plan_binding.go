package orchestrators

import (
	"context"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/google/uuid"
)

type SensorPlanBinding struct {
	dbRepo     ports.SensorPlanBindingRepository
	redisRepo  redisadapter.SensorPlanBinding
	localCache cache.Local[*domain.SensorPlanBinding]
	logger     log.Logger
	asyncPool  ports.AsyncSubmitter
}

var _ ports.SensorPlanBindingOrchestrator = (*SensorPlanBinding)(nil)

func (o *SensorPlanBinding) GetActiveByInstanceID(ctx context.Context, instanceID uuid.UUID) (*domain.SensorPlanBinding, error) {
	cacheKey := cache.FormatActivePlanKey(instanceID.String())
	if spb, ok := o.localCache.Get(cacheKey); ok {
		return spb, nil
	}

	spb, err := o.redisRepo.GetActiveByInstanceID(ctx, instanceID)
	if err == nil {
		return spb, nil
	}
	if !errors.Is(err, errors.ErrNotFound) {
		o.logger.Warn(errors.WrapErr(err))
	}

	spb, err = o.dbRepo.GetActiveByInstanceID(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	o.localCache.Set(cacheKey, spb)

	if redisCli := o.redisRepo.Client(); redisCli.IsAlive() {
		if err := o.asyncPool.Submit(func() {
			if err := o.redisRepo.SetOne(ctx, spb); err != nil {
				o.logger.Warn(errors.WrapErr(err))
			}
		}); err != nil {
			o.logger.Warn(errors.WrapErr(err))
		}
	}

	return spb, nil
}

func (o *SensorPlanBinding) ListActiveByInstanceIDs(ctx context.Context, instanceIDs []uuid.UUID) ([]*domain.SensorPlanBinding, error) {

	return nil, nil
}
