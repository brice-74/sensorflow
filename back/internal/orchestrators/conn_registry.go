package orchestrators

import (
	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
	"github.com/google/uuid"
)

type ClickhouseConnRegistry struct {
}

var _ clickhouse.ConnRegistry = (*ClickhouseConnRegistry)(nil)

func (r *ClickhouseConnRegistry) Get(key uuid.UUID) (*clickhouse.HealthyClient, error) {
	return nil, nil
}
