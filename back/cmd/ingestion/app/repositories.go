package app

import (
	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/adapters/postgres/repo"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	SensorGateway  ports.SensorGatewayRepo
	SensorInstance ports.SensorInstanceRepo
}

func NewRepositories(sqlxDB *sqlx.DB) *Repositories {
	sqlxRepo := &postgres.SqlxRepo{DB: sqlxDB}

	repositories := Repositories{
		SensorGateway:  repo.NewSensorGateway(sqlxRepo),
		SensorInstance: repo.NewSensorInstance(sqlxRepo),
	}

	return &repositories
}
