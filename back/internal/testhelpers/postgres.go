package testhelpers

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer encapsule le conteneur et la connexion DB
type PostgresContainer struct {
	DB        *sql.DB
	container testcontainers.Container
}

// SetupPostgres démarre un conteneur Postgres et retourne PostgresContainer
func SetupPostgres(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17.6",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "secret",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, _ := container.Host(ctx)
	port, _ := container.MappedPort(ctx, "5432")

	dsn := fmt.Sprintf("postgres://postgres:secret@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		container.Terminate(ctx)
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		container.Terminate(ctx)
		return nil, err
	}

	return &PostgresContainer{
		DB:        db,
		container: container,
	}, nil
}

// Teardown ferme la DB et termine le conteneur
func (p *PostgresContainer) Teardown(ctx context.Context) {
	if p.DB != nil {
		p.DB.Close()
	}
	if p.container != nil {
		p.container.Terminate(ctx)
	}
}
