package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/brice-74/sensorflow/pkg/ulid"
	"github.com/redis/go-redis/v9"

	redisadapter "github.com/brice-74/sensorflow/internal/adapters/redis"
)

type SensorGateway struct {
	*redisadapter.Repo
}

var _ ports.SensorGatewayRepo = (*SensorGateway)(nil)

func NewSensorGateway(repo *redisadapter.Repo) *SensorGateway {
	return &SensorGateway{
		Repo: repo,
	}
}

func (*SensorGateway) cmdGetOneByID(ctx context.Context, cmdable redis.Cmdable, ID ulid.ULID) *redis.StringCmd {
	return cmdable.Get(ctx, cache.FormatEntityKey(cache.SensorGatewayKey, ID.String()))
}

func (r *SensorGateway) CmdGetOneByID(ctx context.Context, ID ulid.ULID) *redis.StringCmd {
	return r.cmdGetOneByID(ctx, r.Cmdable(ctx), ID)
}

func (r *SensorGateway) GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	res, err := r.cmdGetOneByID(ctx, r.UnaryCmdable(ctx), ID).Result()
	if err != nil {
		return nil, err
	}
	return gobutil.Decode[domain.SensorGateway]([]byte(res))
}
