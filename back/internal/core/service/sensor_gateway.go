package service

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	repo         ports.SensorGatewayRepo
	instanceRepo ports.SensorInstanceRepo
}

func NewSensorGateway(
	repo ports.SensorGatewayRepo,
	instanceRepo ports.SensorInstanceRepo,
) *SensorGateway {
	return &SensorGateway{
		repo,
		instanceRepo,
	}
}

func (s *SensorGateway) GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (*domain.SensorGateway, error) {
	sensorGateway, err := s.repo.GetOneByID(ctx, gatewayID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	sensorInstances, err := s.instanceRepo.ListByGatewayID(ctx, gatewayID)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	sensorGateway.SensorInstances = sensorInstances
	return sensorGateway, nil
}
