package app

import (
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/adapters/redis"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/jmoiron/sqlx"
)

type PostgresRepositories struct {
	SensorGatewayAuth ports.SensorGatewayAuthRepository
	SensorGateway     ports.SensorGatewayRepository
	SensorInstance    ports.SensorInstanceRepository
}

func NewPostgresRepositories(sqlxDB *sqlx.DB) *PostgresRepositories {
	sqlxRepo := postgres.NewSqlxRepo(sqlxDB)

	repositories := PostgresRepositories{
		SensorGatewayAuth: postgres.NewSensorGatewayAuthRepository(sqlxRepo),
		SensorGateway:     postgres.NewSensorGateway(sqlxRepo),
		SensorInstance:    postgres.NewSensorInstance(sqlxRepo),
	}

	return &repositories
}

type RedisRepositories struct {
	SensorGatewayAuth redis.SensorGatewayAuth
	SensorGateway     redis.SensorGateway
	SensorInstance    redis.SensorInstance
}

func NewRedisRepositories(client *redis.HealthyClient) *RedisRepositories {
	repositories := RedisRepositories{
		SensorGatewayAuth: redis.NewSensorGatewayAuth(client, 10*time.Minute),
		SensorGateway:     redis.NewSensorGateway(client, 10*time.Minute),
		SensorInstance:    redis.NewSensorInstance(client, 7*time.Minute),
	}

	return &repositories
}

type ClickhouseRepositories struct {
	AccelGyroStd *clickhouse.AccelGyroStdRepo
}

func NewClickhouseRepositories(client *clickhouse.HealthyClient) *ClickhouseRepositories {
	repositories := ClickhouseRepositories{
		AccelGyroStd: clickhouse.NewAccelGyroStdRepo(client),
	}

	return &repositories
}
