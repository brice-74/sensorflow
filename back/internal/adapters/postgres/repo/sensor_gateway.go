package repo

import (
	"context"
	"fmt"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/domain"
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

func (r *SensorInstance) GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorGateway, error) {
	var gateway *domain.SensorGateway
	if err := r.
		Executor(ctx).
		SelectContext(
			ctx,
			gateway,
			sensorGatewayGetOneByIDQuery,
			ID,
		); err != nil {
		return nil, fmt.Errorf("failed to get one sensor gateway %s: %w", ID, err)
	}

	return gateway, nil
}
