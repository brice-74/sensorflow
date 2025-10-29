package orchestrators

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	dbRepo            postgres.SensorGateway
	dbInstanceRepo    postgres.SensorInstance
	redisRepo         redisadapter.SensorGateway
	redisInstanceRepo redisadapter.SensorInstance
	localCache        cache.Local[*domain.SensorGateway]
	invalidator       cache.InvalidatorPublisher[string]
	logger            log.Logger
	asyncPool         ports.AsyncSubmitter
}

func (o *SensorGateway) GetOneWithInstances(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	strID := ID.String()
	hydratedKey := cache.FormatHydratedKey(cache.SensorGatewayKey, strID, cache.WithSensorInstancesKey)

	if v, ok := o.localCache.Get(hydratedKey); ok {
		return v, nil
	}

	if redisCli := o.redisRepo.Rdb; redisCli.IsHealthy() {
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

		pipe := redisCli.Pipeline()
		ctxPipe := redisadapter.WithPipeline(ctx, pipe)

		gatewayCmd := o.redisRepo.CmdGetOneByID(ctxPipe, strID)
		idsCmd := o.redisInstanceRepo.CmdListIDsByGatewayID(ctxPipe, strID)

		if _, err := pipe.Exec(ctxPipe); err != nil {
			redisErrors = append(redisErrors, redisCli.HandleError(err))
		}

		if g, err := gatewayCmd.Result(); err == nil {
			gateway = g
			gotGatewayFromRedis = true
		} else {
			if !errors.Is(err, errors.ErrNotFound) {
				gotGatewayFromRedis = true
			} else {
				redisErrors = append(redisErrors, err)
			}
		}

		if ids, err := idsCmd.Result(); err == nil {
			instanceIDs = ids
			gotIDsFromRedis = true
		} else {
			if !errors.Is(err, errors.ErrNotFound) {
				gotIDsFromRedis = true
			} else {
				redisErrors = append(redisErrors, err)
			}
		}

		if gotIDsFromRedis {
			if lenInstanceIDs := len(instanceIDs); lenInstanceIDs > 0 {
				pipe2 := redisCli.Pipeline()
				ctxPipe2 := redisadapter.WithPipeline(ctx, pipe2)

				instancesCmd := o.redisInstanceRepo.CmdGetByIDs(ctxPipe2, instanceIDs)

				if _, err := pipe2.Exec(ctxPipe2); err != nil {
					redisErrors = append(redisErrors, redisCli.HandleError(err))
				}

				if vals, err := instancesCmd.Result(); err == nil {
					instances = vals
					gotInstancesFromRedis = true
				} else {
					if !errors.Is(err, errors.ErrNotFound) {
						gotInstancesFromRedis = true
					} else {
						redisErrors = append(redisErrors, err)
					}
				}
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

			insts, err := o.dbInstanceRepo.ListByGatewayID(ctx, ID)
			if err != nil {
				return nil, errors.WrapErr(err)
			}

			gateway.SensorInstances = insts
			o.localCache.Set(hydratedKey, gateway)
			return gateway, nil
		}
	}

	gwFromDB, err := o.dbRepo.GetOneByID(ctx, ID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	instsFromDB, err := o.dbInstanceRepo.ListByGatewayID(ctx, ID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	gwFromDB.SensorInstances = instsFromDB
	o.localCache.Set(hydratedKey, gwFromDB)

	if err := o.asyncPool.Submit(func() {
		_ = o.RedisRepo.SetOne(ctx, gwFromDB)
		_ = o.RedisInstanceRepo.SetMany(ctx, instsFromDB)
		_ = o.RedisInstanceRepo.SetIDsByGatewayID(ctx, strID, instsFromDB)
	}); err != nil {
		o.logger.Warn(errors.WrapErr(err))
	}

	return gwFromDB, nil
}
