//go:build integration_postgres

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/stretchr/testify/require"
)

func setupTxStateTest(t *testing.T) {
	ctx := context.Background()

	_, err := sqlxDB.ExecContext(ctx, `
		CREATE TABLE test_tx_state (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = sqlxDB.ExecContext(ctx, `DROP TABLE IF EXISTS test_tx_state`)
	})
}

func truncateTxStateTable(t *testing.T) {
	_, err := sqlxDB.Exec(`TRUNCATE TABLE test_tx_state`)
	require.NoError(t, err)
}

func TestSqlxTxState(t *testing.T) {
	setupTxStateTest(t)

	state := postgres.NewSqlxTxState(sqlxDB)

	t.Run("simple commit and rollback transaction", func(t *testing.T) {
		truncateTxStateTable(t)

		t.Run("normal commit", func(t *testing.T) {
			err := state.WithTransaction(context.Background(), nil, func(uow ports.TxUnitOfWork) error {
				tx, ok := postgres.GetSqlxTxFromContext(uow.Context())
				require.True(t, ok)

				_, err := tx.ExecContext(context.Background(),
					`INSERT INTO test_tx_state(value) VALUES($1)`, "commit")
				return err
			})
			require.NoError(t, err)

			var count int
			err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_state WHERE value = $1`, "commit")
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("rollback on error", func(t *testing.T) {
			err := state.WithTransaction(context.Background(), nil, func(uow ports.TxUnitOfWork) error {
				tx, ok := postgres.GetSqlxTxFromContext(uow.Context())
				require.True(t, ok)

				_, err := tx.ExecContext(context.Background(),
					`INSERT INTO test_tx_state(value) VALUES($1)`, "rollback")
				require.NoError(t, err)

				return errors.New("")
			})
			require.Error(t, err)

			var count int
			err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_state WHERE value = $1`, "rollback")
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})
	})

	t.Run("nested savepoints with intermediate rollback", func(t *testing.T) {
		truncateTxStateTable(t)

		err := state.WithTransaction(context.Background(), nil, func(uow ports.TxUnitOfWork) error {
			txMain := uow.(*postgres.SqlxTxState).Tx()

			_, err := txMain.Exec(`INSERT INTO test_tx_state(value) VALUES($1)`, "main")
			require.NoError(t, err)

			// Savepoint 1
			uow1, err := uow.Transaction(uow.Context(), nil)
			require.NoError(t, err)
			save1 := uow1.(*postgres.SqlxTxState).Tx()
			_, err = save1.Exec(`INSERT INTO test_tx_state(value) VALUES($1)`, "save1")
			require.NoError(t, err)

			// Savepoint 2
			uow2, err := uow.Transaction(uow1.Context(), nil)
			require.NoError(t, err)
			save2 := uow2.(*postgres.SqlxTxState).Tx()
			_, err = save2.Exec(`INSERT INTO test_tx_state(value) VALUES($1)`, "save2")
			require.NoError(t, err)

			// Revert savepoint 2
			require.NoError(t, uow2.Revert(context.Background()))
			return nil
		})
		require.NoError(t, err)

		var count int
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_state WHERE value = $1`, "save2")
		require.NoError(t, err)
		require.Equal(t, 0, count)
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_state WHERE value = $1`, "save1")
		require.NoError(t, err)
		require.Equal(t, 1, count)
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_state WHERE value = $1`, "main")
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("finish et revert edge cases", func(t *testing.T) {
		truncateTxStateTable(t)
		t.Run("finish without transaction", func(t *testing.T) {
			require.Error(t, state.Finish(context.Background()))
		})
		t.Run("revert without transaction", func(t *testing.T) {
			require.Error(t, state.Revert(context.Background()))
		})
	})

	t.Run("shared context between rest", func(t *testing.T) {
		truncateTxStateTable(t)
		err := state.WithTransaction(context.Background(), nil, func(tx ports.TxUnitOfWork) error {
			tx1, ok1 := postgres.GetSqlxTxFromContext(tx.Context())
			tx2, ok2 := postgres.GetSqlxTxFromContext(tx.Context())
			require.True(t, ok1)
			require.True(t, ok2)
			require.Same(t, tx1, tx2)
			return nil
		})
		require.NoError(t, err)
	})
}
