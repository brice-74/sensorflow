//go:build integration_postgres

package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/brice-74/sensorflow/internal/testhelpers"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := testhelpers.SetupPostgres(ctx)
	if err != nil {
		panic(err)
	}
	defer pg.Teardown(ctx)

	testDB = pg.DB

	os.Exit(m.Run())
}
