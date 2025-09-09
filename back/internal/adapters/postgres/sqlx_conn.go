package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxConnCtxKey struct{}

func GetSqlxConnFromContext(ctx context.Context) (*sqlx.Conn, bool) {
	conn, ok := ctx.Value(sqlxConnCtxKey{}).(*sqlx.Conn)
	return conn, ok
}

type SqlxConnState struct {
	db *sqlx.DB
}

// WithConnection acquires a connection and executes the given function with it.
// The connection is released after the function returns, even in case of error or panic.
func (s *SqlxConnState) WithConnection(ctx context.Context, fn func(ctx context.Context) error) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx = context.WithValue(ctx, sqlxConnCtxKey{}, conn)
	return fn(ctx)
}

// Connection returns the current connection and a release function.
// Caller must call releaseFn() when done.
func (s *SqlxConnState) Connection(ctx context.Context) (func() error, context.Context, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, nil, err
	}

	releaseFn := func() error {
		return conn.Close()
	}

	ctx = context.WithValue(ctx, sqlxConnCtxKey{}, conn)
	return releaseFn, ctx, nil
}
