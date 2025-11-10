package app

import (
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/jmoiron/sqlx"
)

type PostgresRepositories struct {
	SensorGateway  ports.SensorGatewayRepository
	SensorInstance ports.SensorInstanceRepository
}

func NewPostgresRepositories(sqlxDB *sqlx.DB) *PostgresRepositories {
	sqlxRepo := postgres.NewSqlxRepo(sqlxDB)

	repositories := PostgresRepositories{
		SensorGateway:  postgres.NewSensorGateway(sqlxRepo),
		SensorInstance: postgres.NewSensorInstance(sqlxRepo),
	}

	return &repositories
}

type RedisRepositories struct {
	SensorGateway  redis.SensorGateway
	SensorInstance redis.SensorInstance
}

func NewRedisRepositories(client *redis.HealthyClient) *PostgresRepositories {

	repositories := PostgresRepositories{
		SensorGateway:  redis.NewSensorGateway(client, 10*time.Minute),
		SensorInstance: redis.NewSensorInstance(client, 7*time.Minute),
	}

	return &repositories
}
