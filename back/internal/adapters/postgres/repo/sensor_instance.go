package repo

import (
	"context"
	"fmt"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	*postgres.SqlxRepo
}

func NewSensorInstance(repo *postgres.SqlxRepo) *SensorInstance {
	return &SensorInstance{repo}
}

const sensorInstanceListByGatewayIDQuery = `
	SELECT * FROM sensor_instance
	WHERE sensor_gateway_id = $1 AND deleted_at IS NULL
`

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*domain.SensorInstance, error) {
	var sensors []*domain.SensorInstance
	if err := r.
		Executor(ctx).
		SelectContext(
			ctx,
			&sensors,
			sensorInstanceListByGatewayIDQuery,
			gatewayID,
		); err != nil {
		return nil, fmt.Errorf("failed to list sensor instances for gateway %s: %w", gatewayID, err)
	}

	return sensors, nil
}
