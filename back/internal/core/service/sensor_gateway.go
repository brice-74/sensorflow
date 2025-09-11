package service

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	uowConn      ports.ConnUnitOfWork
	repo         ports.SensorGatewayRepo
	instanceRepo ports.SensorInstanceRepo
}

func NewSensorGateway(
	uowConn ports.ConnUnitOfWork,
	repo ports.SensorGatewayRepo,
	instanceRepo ports.SensorInstanceRepo,
) *SensorGateway {
	return &SensorGateway{
		uowConn,
		repo,
		instanceRepo,
	}
}

func (s *SensorGateway) GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (sensorGateway *domain.SensorGateway, err error) {
	err = s.uowConn.WithConnection(ctx, func(ctx context.Context) error {
		var e error

		sensorGateway, e = s.repo.GetOneByID(ctx, gatewayID)
		if e != nil {
			return errors.WrapErr(e)
		}

		sensorGateway.SensorInstances, e = s.instanceRepo.ListByGatewayID(ctx, gatewayID)
		if e != nil {
			return errors.WrapErr(e)
		}

		return nil
	})
	return
}
