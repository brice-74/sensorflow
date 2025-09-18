//go:build integration_postgres

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/brice-74/sensorflow/internal/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sqlxDB *sqlx.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := testhelpers.SetupPostgres(ctx)
	if err != nil {
		panic(err)
	}
	defer pg.Teardown(ctx)

	sqlxDB = pg.SqlxDB()

	os.Exit(m.Run())
}
