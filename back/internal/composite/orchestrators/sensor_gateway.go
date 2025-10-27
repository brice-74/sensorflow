package orchestrators

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	DBRepo            postgres.SensorGateway
	DBInstanceRepo    postgres.SensorInstance
	RedisRepo         redisadapter.SensorGateway
	RedisInstanceRepo redisadapter.SensorInstance
	LocalCache        cache.Local[*domain.SensorGateway]
	Invalidator       cache.InvalidatorPublisher[string]
}

func (o *SensorGateway) GetOneWithInstances(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	strID := ID.String()
	if v, found := o.LocalCache.Get(
		cache.FormatHydratedKey(cache.SensorGatewayKey, strID, cache.WithSensorInstancesKey),
	); found {
		return v, nil
	}

	var (
		sensorInstanceIds []string
		sensorgateway     *domain.SensorGateway
		errs              []error
		err               error
	)

	pipe := o.RedisRepo.Client.Pipeline()
	ctx = redisadapter.WithPipeline(ctx, pipe)

	sensorGatewayCmd := o.RedisRepo.CmdGetOneByID(ctx, strID)
	sensorInstancesIdsCmd := o.RedisInstanceRepo.CmdListIDsByGatewayID(ctx, strID)
	if _, err := pipe.Exec(ctx); err != nil {
		errs = append(errs, redisadapter.HandleError(err))
	}

	sensorgateway, err = sensorGatewayCmd.Result()
	if err != nil {

	}
	sensorInstanceIds, err = sensorInstancesIdsCmd.Result()
	if err != nil {

	}

	// si on a bien la liste d'ids
	sensorInstancesCmd := o.RedisInstanceRepo.CmdGetByIDs(ctx, sensorInstanceIds)
	if _, err := pipe.Exec(ctx); err != nil {
		errs = append(errs, redisadapter.HandleError(err))
	}
	sensorgateway.SensorInstances, err = sensorInstancesCmd.Result()
	if err != nil {

	}
	// si on a bienj les insatnces on retourne

	// si sensor gateway et insatnces ne sont pas trouver depuis les cache, on lance une connexion sql commune pour faire passer les deux requetes sur la meme

	// si on a pas le gateway on va en db
	sensorgateway, err = o.DBRepo.GetOneByID(ctx, ID)
	if err != nil {
		// on aura une erreru et le sensorgateway nil
	}

	// si on a pas les instances
	sensorgateway.SensorInstances, err = o.DBInstanceRepo.ListByGatewayID(ctx, ID)
	if err != nil {
		// on aura une erreru et le sensorgateway nil
	}

	return nil, nil
}
