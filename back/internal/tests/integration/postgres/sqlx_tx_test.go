//go:build integration_postgres

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/stretchr/testify/require"
)

func setupTxManagerTest(t *testing.T) {
	ctx := context.Background()

	_, err := sqlxDB.ExecContext(ctx, `
		CREATE TABLE test_tx_manager (
			id SERIAL PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = sqlxDB.ExecContext(ctx, `DROP TABLE IF EXISTS test_tx_manager`)
	})
}

func truncateTxManagerTable(t *testing.T) {
	_, err := sqlxDB.Exec(`TRUNCATE TABLE test_tx_manager`)
	require.NoError(t, err)
}

func TestSqlxTxManager(t *testing.T) {
	setupTxManagerTest(t)

	m := postgres.NewSqlxTxManager(sqlxDB)

	t.Run("simple commit and rollback transaction", func(t *testing.T) {
		truncateTxManagerTable(t)

		t.Run("normal commit", func(t *testing.T) {
			err := m.WithTransaction(context.Background(), nil, func(ctx context.Context) error {
				tx, ok := postgres.GetSqlxTx(ctx)
				require.True(t, ok)

				_, err := tx.ExecContext(context.Background(),
					`INSERT INTO test_tx_manager(value) VALUES($1)`, "commit")
				return err
			})
			require.NoError(t, err)

			var count int
			err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_manager WHERE value = $1`, "commit")
			require.NoError(t, err)
			require.Equal(t, 1, count)
		})

		t.Run("rollback on error", func(t *testing.T) {
			err := m.WithTransaction(context.Background(), nil, func(ctx context.Context) error {
				tx, ok := postgres.GetSqlxTx(ctx)
				require.True(t, ok)

				_, err := tx.ExecContext(context.Background(),
					`INSERT INTO test_tx_manager(value) VALUES($1)`, "rollback")
				require.NoError(t, err)

				return errors.New("")
			})
			require.Error(t, err)

			var count int
			err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_manager WHERE value = $1`, "rollback")
			require.NoError(t, err)
			require.Equal(t, 0, count)
		})
	})

	t.Run("nested savepoints with intermediate rollback", func(t *testing.T) {
		truncateTxManagerTable(t)

		err := m.WithTransaction(context.Background(), nil, func(ctx context.Context) error {
			txMain, _ := postgres.GetSqlxTx(ctx)

			_, err := txMain.Exec(`INSERT INTO test_tx_manager(value) VALUES($1)`, "main")
			require.NoError(t, err)

			// Savepoint 1
			sub1, err := m.Begin(ctx, nil)
			require.NoError(t, err)
			_, err = sub1.Tx().Exec(`INSERT INTO test_tx_manager(value) VALUES($1)`, "save1")
			require.NoError(t, err)

			// Savepoint 2
			sub2, err := m.Begin(sub1.Context(), nil)
			require.NoError(t, err)
			_, err = sub2.Tx().Exec(`INSERT INTO test_tx_manager(value) VALUES($1)`, "save2")
			require.NoError(t, err)

			// Revert savepoint 2
			require.NoError(t, sub2.Rollback(context.Background()))
			return nil
		})
		require.NoError(t, err)

		var count int
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_manager WHERE value = $1`, "save2")
		require.NoError(t, err)
		require.Equal(t, 0, count)
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_manager WHERE value = $1`, "save1")
		require.NoError(t, err)
		require.Equal(t, 1, count)
		err = sqlxDB.Get(&count, `SELECT COUNT(*) FROM test_tx_manager WHERE value = $1`, "main")
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("finish et revert edge cases", func(t *testing.T) {
		truncateTxManagerTable(t)
		t.Run("finish without transaction", func(t *testing.T) {
			require.Error(t, m.Commit(context.Background()))
		})
		t.Run("revert without transaction", func(t *testing.T) {
			require.Error(t, m.Rollback(context.Background()))
		})
	})

	t.Run("shared context between rest", func(t *testing.T) {
		truncateTxManagerTable(t)
		err := m.WithTransaction(context.Background(), nil, func(ctx context.Context) error {
			tx1, ok1 := postgres.GetSqlxTx(ctx)
			tx2, ok2 := postgres.GetSqlxTx(ctx)
			require.True(t, ok1)
			require.True(t, ok2)
			require.Same(t, tx1, tx2)
			return nil
		})
		require.NoError(t, err)
	})
}
