package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	*postgres.SqlxRepo
}

var _ ports.SensorInstanceRepo = (*SensorInstance)(nil)

func NewSensorInstance(repo *postgres.SqlxRepo) *SensorInstance {
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
	if err = postgres.HandleSelectError(err); err != nil {
		return nil, errors.WrapErr(err)
	}

	return sensors, nil
}
