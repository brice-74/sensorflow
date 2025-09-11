package repo

import (
	"context"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorGateway struct {
	*postgres.SqlxRepo
}

func NewSensorGateway(repo *postgres.SqlxRepo) *SensorGateway {
	return &SensorGateway{repo}
}

const sensorGatewayGetOneByIDQuery = `
	SELECT * FROM sensor_gateway
	WHERE id = $1 AND deleted_at IS NULL
	LIMIT 1
`

func (r *SensorGateway) GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	var gateway *domain.SensorGateway
	if err := r.
		Executor(ctx).
		SelectContext(
			ctx,
			gateway,
			sensorGatewayGetOneByIDQuery,
			ID,
		); err != nil {
		return nil, errors.Wrap(err, "failed to get sensor gateway")
	}

	return gateway, nil
}
