package postgres

import (
	"context"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/jmoiron/sqlx"
)

type sqlxConnCtxKeyType struct{}

var sqlxConnCtxKey = sqlxConnCtxKeyType{}

func GetSqlxConn(ctx context.Context) (*sqlx.Conn, bool) {
	conn, ok := ctx.Value(sqlxConnCtxKey).(*sqlx.Conn)
	return conn, ok
}

type SqlxConnManager struct {
	db *sqlx.DB
}

func NewSqlxConnManager(db *sqlx.DB) *SqlxConnManager {
	if db == nil {
		panic("*sqlx.DB cannot be nil")
	}
	return &SqlxConnManager{db}
}

// WithConnection acquires a connection and executes the given function with it.
// The connection is released after the function returns, even in case of error or panic.
func (s *SqlxConnManager) WithConnection(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return errors.Wrap(err, "get *sqlx.Conn")
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			err = errors.Join(err, errors.Wrap(cerr, "close *sql.Conn"))
		}
	}()

	ctx = context.WithValue(ctx, sqlxConnCtxKey, conn)
	return fn(ctx)
}

// Connection returns a new connection has context value and a release function.
// Caller must call releaseFn() when done.
func (s *SqlxConnManager) Connection(ctx context.Context) (func() error, context.Context, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "get *sqlx.Conn")
	}

	releaseFn := func() error {
		return conn.Close()
	}

	ctx = context.WithValue(ctx, sqlxConnCtxKey, conn)
	return releaseFn, ctx, nil
}
