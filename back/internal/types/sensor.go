package types

import (
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/google/uuid"
)

type SensorGatewayAuthPlanRow struct {
	TenantID         uuid.UUID             `db:"tenant_id"`
	SensorInstanceID uuid.UUID             `db:"sensor_instance_id"`
	ActivePlan       domain.SensorPlanType `db:"active_plan"`
}
