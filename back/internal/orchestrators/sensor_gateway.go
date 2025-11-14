package orchestrators

import (
	"context"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	dbRepo            ports.SensorGatewayRepository
	dbInstanceRepo    ports.SensorInstanceRepository
	redisRepo         redisadapter.SensorGateway
	redisInstanceRepo redisadapter.SensorInstance
	localCache        cache.Local[*domain.SensorGateway]
	logger            log.Logger
	asyncPool         ports.AsyncSubmitter
}

func NewSensorGateway(
	dbRepo ports.SensorGatewayRepository,
	dbInstanceRepo ports.SensorInstanceRepository,
	redisRepo redisadapter.SensorGateway,
	redisInstanceRepo redisadapter.SensorInstance,
	localCache cache.Local[*domain.SensorGateway],
	logger log.Logger,
	asyncPool ports.AsyncSubmitter,
) *SensorGateway {
	return &SensorGateway{
		dbRepo:            dbRepo,
		dbInstanceRepo:    dbInstanceRepo,
		redisRepo:         redisRepo,
		redisInstanceRepo: redisInstanceRepo,
		localCache:        localCache,
		logger:            logger,
		asyncPool:         asyncPool,
	}
}

func (o *SensorGateway) GetOneWithInstances(ctx context.Context, id ulid.ULID) (*domain.SensorGateway, error) {
	strID := id.String()
	hydratedKey := cache.FormatHydratedKey(cache.SensorGatewayKey, strID, cache.WithSensorInstancesKey)

	if v, ok := o.localCache.Get(hydratedKey); ok {
		return v, nil
	}

	if redisCli := o.redisRepo.Client(); redisCli.IsHealthy() {
		var (
			redisErrors []error

			gateway             *domain.SensorGateway
			gotGatewayFromRedis bool

			instanceIDs     []string
			gotIDsFromRedis bool

			instances             []*domain.SensorInstance
			gotInstancesFromRedis bool
		)
		defer func() {
			if len(redisErrors) > 0 {
				o.logger.Warn(errors.JoinWrap(redisErrors...))
			}
		}()

		pipe, ctxPipe := redisCli.NewPipeline(ctx)

		gatewayCmd := o.redisRepo.CmdGetOneByStrID(ctxPipe, strID)
		idsCmd := o.redisInstanceRepo.CmdListIDsByGatewayID(ctxPipe, strID)

		if _, err := pipe.Exec(ctxPipe); err != nil {
			redisErrors = append(redisErrors, redisCli.HandleError(err))
		}

		if g, err := gatewayCmd.Result(); err == nil {
			gateway = g
			gotGatewayFromRedis = true
		} else {
			if errors.Is(err, errors.ErrNotFound) {
				gotGatewayFromRedis = true
			} else {
				redisErrors = append(redisErrors, err)
			}
		}

		if ids, err := idsCmd.Result(); err == nil {
			instanceIDs = ids
			gotIDsFromRedis = true
		} else {
			if errors.Is(err, errors.ErrNotFound) {
				gotIDsFromRedis = true
			} else {
				redisErrors = append(redisErrors, err)
			}
		}

		if gotIDsFromRedis {
			if lenInstanceIDs := len(instanceIDs); lenInstanceIDs > 0 {
				insts, err := o.redisInstanceRepo.GetManyByStrIDs(ctx, instanceIDs)
				if err != nil {
					if errors.Is(err, errors.ErrNotFound) {
						gotInstancesFromRedis = true
					} else {
						redisErrors = append(redisErrors, err)
					}
				}

				instances = insts
			} else if lenInstanceIDs == 0 {
				gotInstancesFromRedis = true
				instances = nil
			}
		}

		if gotGatewayFromRedis {
			if gotInstancesFromRedis {
				gateway.SensorInstances = instances
				o.localCache.Set(hydratedKey, gateway)
				return gateway, nil
			}

			insts, err := o.dbInstanceRepo.ListByGatewayID(ctx, id)
			if err != nil {
				return nil, errors.WrapErr(err)
			}

			gateway.SensorInstances = insts
			o.localCache.Set(hydratedKey, gateway)
			return gateway, nil
		}
	}

	gwFromDB, err := o.dbRepo.GetOneByID(ctx, id)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	instsFromDB, err := o.dbInstanceRepo.ListByGatewayID(ctx, id)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	gwFromDB.SensorInstances = instsFromDB
	o.localCache.Set(hydratedKey, gwFromDB)

	if redisCli := o.redisRepo.Client(); redisCli.IsHealthy() {
		if err := o.asyncPool.Submit(func() {
			var redisErrors []error
			defer func() {
				if len(redisErrors) > 0 {
					o.logger.Warn(errors.JoinWrap(redisErrors...))
				}
			}()

			pipe, ctxPipe := redisCli.NewPipeline(context.Background())

			setGtwCmd := o.redisRepo.CmdSetOne(ctxPipe, gwFromDB)
			setInstsCmd, setInstsTtlCmds := o.redisInstanceRepo.CmdSetMany(ctxPipe, instsFromDB)

			if _, err := pipe.Exec(ctxPipe); err != nil {
				redisErrors = append(redisErrors, redisCli.HandleError(err))
			}

			if _, err := setGtwCmd.Result(); err != nil {
				redisErrors = append(redisErrors, err)
			}
			if _, err = setInstsCmd.Result(); err != nil {
				redisErrors = append(redisErrors, err)
			}
			if _, err = setInstsTtlCmds.Result(); err != nil {
				redisErrors = append(redisErrors, err)
			}

			if len(redisErrors) == 0 {
				if err = o.redisInstanceRepo.SetIDsByGatewayID(ctx, gwFromDB.ID, instsFromDB); err != nil {
					redisErrors = append(redisErrors, err)
				}
			}

		}); err != nil {
			o.logger.Warn(errors.WrapErr(err))
		}
	}

	return gwFromDB, nil
}
