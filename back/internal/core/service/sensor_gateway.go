package service

import (
	"context"
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	uowConn      ports.ConnUnitOfWork
	repo         ports.SensorGatewayRepo
	instanceRepo ports.SensorInstanceRepo
}

func (s *SensorGateway) GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (*domain.SensorGateway, error) {
	release, ctx, err := s.uowConn.Connection(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get one sensor gateway %s: %w", ID, err)
	}
}
