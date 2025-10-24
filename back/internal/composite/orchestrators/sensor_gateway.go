package orchestrators

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	local
}

func (s *SensorGateway) GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (*domain.SensorGateway, error) {
	ctx = redis.WithConn(ctx, s.repo.RedisRepo.Client.Conn())
	close, ctx, err := postgres.Connection(ctx, s.repo.DBRepo.DB)
	if err != nil {
		return nil, err
	}
	defer close()

	sensorGateway, err := s.repo.GetOneByID(ctx, gatewayID)
	if err != nil {
		return nil, err
	}

	sensorInstances, err := s.repoInstances.ListByGatewayID(ctx, gatewayID)
	if err != nil {
		return nil, err
	}

	sensorGateway.SensorInstances = sensorInstances

	return sensorGateway, nil
}
