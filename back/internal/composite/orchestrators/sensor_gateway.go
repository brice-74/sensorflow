package orchestrators

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	DBRepo            postgres.SensorGateway
	DBInstanceRepo    postgres.SensorInstance
	RedisRepo         redisadapter.SensorGateway
	RedisInstanceRepo redisadapter.SensorInstance
	LocalCache        cache.Local[*domain.SensorGateway]
	Invalidator       cache.InvalidatorPublisher[string]
	Logger            log.Logger
}

func (o *SensorGateway) GetOneWithInstances(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	strID := ID.String()
	hydratedKey := cache.FormatHydratedKey(cache.SensorGatewayKey, strID, cache.WithSensorInstancesKey)

	if v, ok := o.LocalCache.Get(hydratedKey); ok {
		return v, nil
	}

	var (
		redisErrors []error

		gateway             *domain.SensorGateway
		gotGatewayFromRedis bool

		instanceIDs     []string
		gotIDsFromRedis bool

		instances             []*domain.SensorInstance
		gotInstancesFromRedis bool
	)

	pipe := o.RedisRepo.Client.Pipeline()
	ctxPipe := redisadapter.WithPipeline(ctx, pipe)

	gatewayCmd := o.RedisRepo.CmdGetOneByID(ctxPipe, strID)
	idsCmd := o.RedisInstanceRepo.CmdListIDsByGatewayID(ctxPipe, strID)

	if _, err := pipe.Exec(ctxPipe); err != nil {
		redisErrors = append(redisErrors, redisadapter.HandleError(err))
	}

	if g, err := gatewayCmd.Result(); err == nil {
		gateway = g
		gotGatewayFromRedis = true
	} else {
		if appErr := redisadapter.HandleError(err); appErr != nil {
			if errors.Is(appErr, errors.ErrNotFound) {
				gotGatewayFromRedis = false
			} else {
				redisErrors = append(redisErrors, appErr)
			}
		} else {
			gotGatewayFromRedis = false
		}
	}

	if ids, err := idsCmd.Result(); err == nil {
		instanceIDs = ids
		gotIDsFromRedis = true
	} else {
		if appErr := redisadapter.HandleError(err); appErr != nil {
			if errors.Is(appErr, errors.ErrNotFound) {
				gotIDsFromRedis = false
			} else {
				redisErrors = append(redisErrors, appErr)
			}
		} else {
			gotIDsFromRedis = false
		}
	}

	if gotIDsFromRedis && len(instanceIDs) > 0 {
		pipe2 := o.RedisInstanceRepo.Client.Pipeline()
		ctxPipe2 := redisadapter.WithPipeline(ctx, pipe2)

		instancesCmd := o.RedisInstanceRepo.CmdGetByIDs(ctxPipe2, instanceIDs)

		if _, err := pipe2.Exec(ctxPipe2); err != nil {
			redisErrors = append(redisErrors, redisadapter.HandleError(err))
		}

		if vals, err := instancesCmd.Result(); err == nil {
			instances = vals
			gotInstancesFromRedis = true
		} else {
			if appErr := redisadapter.HandleError(err); appErr != nil {
				if errors.Is(appErr, errors.ErrNotFound) {
					gotInstancesFromRedis = false
				} else {
					redisErrors = append(redisErrors, appErr)
				}
			} else {
				gotInstancesFromRedis = false
			}
		}
	} else if gotIDsFromRedis && len(instanceIDs) == 0 {
		gotInstancesFromRedis = true
		instances = nil
	}

	if gotGatewayFromRedis {
		if gotInstancesFromRedis {
			gateway.SensorInstances = instances
			o.LocalCache.Set(hydratedKey, gateway)

			// Log redis warnings but don't fail the call
			/* if len(redisErrors) > 0 {
				o.Logger.Warn("partial redis errors while fetching gateway",
					log.Tags{"gateway_id": strID, "warnings": redisErrors})
			} */
			return gateway, nil
		}

		insts, err := o.DBInstanceRepo.ListByGatewayID(ctx, ID)
		if err != nil {
			return nil, errors.WrapErr(err)
		}

		gateway.SensorInstances = insts
		o.LocalCache.Set(hydratedKey, gateway)

		// async: repopulate redis (non-blocking)
		/* o.safeAsync("redis_repopulate_after_db_instances", func() {
			_ = o.RedisInstanceRepo.SetMany(ctx, insts)
			_ = o.RedisInstanceRepo.SetIDsByGatewayID(ctx, strID, insts)
			_ = o.RedisRepo.SetOne(ctx, gateway)
		}) */
		return gateway, nil
	}

	gwFromDB, err := o.DBRepo.GetOneByID(ctx, ID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	instsFromDB, err := o.DBInstanceRepo.ListByGatewayID(ctx, ID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	gwFromDB.SensorInstances = instsFromDB
	o.LocalCache.Set(hydratedKey, gwFromDB)

	// async cache population
	/* o.safeAsync("redis_populate_after_db_fetch", func() {

		_ = o.RedisRepo.SetOne(ctx, gwFromDB)
		_ = o.RedisInstanceRepo.SetMany(ctx, instsFromDB)
		_ = o.RedisInstanceRepo.SetIDsByGatewayID(ctx, strID, instsFromDB)
	}) */

	return gwFromDB, nil
}
