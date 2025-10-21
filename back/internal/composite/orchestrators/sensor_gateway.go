package orchestrators

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
}

func (s *SensorGateway) GetOneWithInstances(ctx context.Context, gatewayID ulid.ULID) (*domain.SensorGateway, error) {

}
