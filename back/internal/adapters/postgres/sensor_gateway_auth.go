package postgres

import (
	"context"

	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/internal/types"
	"github.com/google/uuid"
)

type SensorGatewayAuthRepository struct {
	*sqlxRepo
}

var _ ports.SensorGatewayAuthRepository = (*SensorGatewayAuthRepository)(nil)

func NewSensorGatewayAuthRepository(repo *sqlxRepo) *SensorGatewayAuthRepository {
	return &SensorGatewayAuthRepository{repo}
}

const sensorGatewayAuthGetPlanRowsByGatewayIDQuery = `
	SELECT
		sg.tenant_id AS tenant_id,
		si.id AS sensor_instance_id,
		sp.plan AS active_plan
	FROM sensor_gateway sg
	LEFT JOIN sensor_instance si
		ON si.sensor_gateway_id = sg.id
		AND si.deleted_at IS NULL
	LEFT JOIN sensor_plan_binding spb
		ON spb.sensor_instance_id = si.id
		AND spb.deleted_at IS NULL
	LEFT JOIN sensor_plan sp
		ON sp.id = spb.sensor_plan_id
		AND sp.deleted_at IS NULL
	WHERE sg.id = $1
		AND sg.deleted_at IS NULL;
`

func (r *SensorGatewayAuthRepository) GetAuthPlanRowsByGatewayID(ctx context.Context, gatewayID uuid.UUID) ([]*types.SensorGatewayAuthPlanRow, error) {
	var sensors []*types.SensorGatewayAuthPlanRow
	err := r.
		ExecutorFromCtx(ctx).
		SelectContext(
			ctx,
			&sensors,
			sensorGatewayAuthGetPlanRowsByGatewayIDQuery,
			gatewayID,
		)
	if err = handleSelectError(err); err != nil {
		return nil, err
	}

	return sensors, nil
}
