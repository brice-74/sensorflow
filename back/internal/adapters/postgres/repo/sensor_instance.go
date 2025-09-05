package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	Client postgres.Client
}

func NewSensorInstance(client postgres.Client) *SensorInstance {
	return &SensorInstance{client}
}

const sensorInstanceListByGatewayIDQuery = `
	SELECT * FROM sensor_instance
	WHERE sensor_gateway_id = $1 AND deleted_at IS NULL
`

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gatewayID ulid.ULID) ([]*SensorInstance, error) {
	var sensors []*SensorInstance
	if err := r.Client.Sqlx().SelectContext(ctx, &sensors, sensorInstanceListByGatewayIDQuery, gatewayID); err != nil {
		return nil, err
	}

	return sensors, nil
}
