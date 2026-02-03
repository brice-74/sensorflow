package postgres

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/google/uuid"
)

type SensorGateway struct {
	*sqlxRepo
}

var _ ports.SensorGatewayRepository = (*SensorGateway)(nil)

func NewSensorGateway(repo *sqlxRepo) *SensorGateway {
	return &SensorGateway{repo}
}

const sensorGatewayGetOneByIDQuery = `
	SELECT * FROM sensor_gateway
	WHERE id = $1 AND deleted_at IS NULL
	LIMIT 1
`

func (r *SensorGateway) GetOneByID(ctx context.Context, ID uuid.UUID) (*domain.SensorGateway, error) {
	var gateway *domain.SensorGateway
	err := r.
		ExecutorFromCtx(ctx).
		SelectContext(
			ctx,
			gateway,
			sensorGatewayGetOneByIDQuery,
			ID,
		)
	if err = handleSelectError(err); err != nil {
		return nil, err
	}

	return gateway, nil
}
