//go:build integration_postgres

package postgres_test

import (
	"context"
	"testing"

	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/stretchr/testify/require"
)

type customError struct {
	msg string
}

func (c *customError) Error() string { return c.msg }

func TestWithConnection_ContextInjectionAndErrorPropagation(t *testing.T) {
	m := postgres.NewSqlxConnManager(sqlxDB)

	t.Run("connection is injected into context", func(t *testing.T) {
		err := m.WithConnection(context.Background(), func(ctx context.Context) error {
			conn, ok := postgres.GetSqlxConn(ctx)
			require.True(t, ok)
			require.NotNil(t, conn)
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("error from function is propagated", func(t *testing.T) {
		wantErr := &customError{"boom"}
		err := m.WithConnection(context.Background(), func(ctx context.Context) error {
			return wantErr
		})
		require.ErrorIs(t, err, wantErr)
	})
}

func TestConnection_InjectionAndReleaseFn(t *testing.T) {
	m := postgres.NewSqlxConnManager(sqlxDB)

	t.Run("connection injected and release works", func(t *testing.T) {
		releaseFn, ctx, err := m.Connection(context.Background())
		require.NoError(t, err)
		require.NotNil(t, releaseFn)

		conn, ok := postgres.GetSqlxConn(ctx)
		require.True(t, ok)
		require.NotNil(t, conn)

		require.NoError(t, releaseFn())
	})

	t.Run("two calls produce distinct connections", func(t *testing.T) {
		release1, ctx1, err1 := m.Connection(context.Background())
		require.NoError(t, err1)
		defer release1()

		release2, ctx2, err2 := m.Connection(context.Background())
		require.NoError(t, err2)
		defer release2()

		conn1, ok1 := postgres.GetSqlxConn(ctx1)
		conn2, ok2 := postgres.GetSqlxConn(ctx2)
		require.True(t, ok1)
		require.True(t, ok2)
		require.NotSame(t, conn1, conn2)
	})
}

func TestSharedConnectionAcrossRepos(t *testing.T) {
	m := postgres.NewSqlxConnManager(sqlxDB)

	err := m.WithConnection(context.Background(), func(ctx context.Context) error {
		// simule deux "repos" utilisant le même context
		conn1, ok1 := postgres.GetSqlxConn(ctx)
		conn2, ok2 := postgres.GetSqlxConn(ctx)
		require.True(t, ok1)
		require.True(t, ok2)
		require.Same(t, conn1, conn2)
		return nil
	})
	require.NoError(t, err)
}

func TestSimpleQuery(t *testing.T) {
	m := postgres.NewSqlxConnManager(sqlxDB)

	err := m.WithConnection(context.Background(), func(ctx context.Context) error {
		conn, ok := postgres.GetSqlxConn(ctx)
		require.True(t, ok)
		var one int
		err := conn.QueryRowxContext(ctx, "SELECT 1").Scan(&one)
		require.NoError(t, err)
		require.Equal(t, 1, one)
		return nil
	})
	require.NoError(t, err)
}
