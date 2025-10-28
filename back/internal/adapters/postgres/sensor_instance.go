package postgres

import (
	"context"

	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	*SqlxRepo
}

var _ ports.SensorInstanceRepo = (*SensorInstance)(nil)

func NewSensorInstance(repo *SqlxRepo) *SensorInstance {
	return &SensorInstance{repo}
}

const sensorInstanceListByGatewayIDQuery = `
	SELECT * FROM sensor_instance
	WHERE sensor_gateway_id = $1 AND deleted_at IS NULL
`

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	var sensors []*domain.SensorInstance
	err := r.
		Executor(ctx).
		SelectContext(
			ctx,
			&sensors,
			sensorInstanceListByGatewayIDQuery,
			gatewayID,
		)
	if err = handleSelectError(err); err != nil {
		return nil, err
	}

	return sensors, nil
}

const sensorInstanceGetOneByIDQuery = `
	SELECT * FROM sensor_instance
	WHERE id = $1 AND deleted_at IS NULL
	LIMIT 1
`

func (r *SensorInstance) GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorInstance, error) {
	var gateway *domain.SensorInstance
	err := r.
		Executor(ctx).
		SelectContext(
			ctx,
			gateway,
			sensorInstanceGetOneByIDQuery,
			ID,
		)
	if err = handleSelectError(err); err != nil {
		return nil, err
	}

	return gateway, nil
}
