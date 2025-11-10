package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type SqlxTxWrapper struct {
	*sqlx.Tx
}

func (tx *SqlxTxWrapper) NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error) {
	return sqlx.NamedQueryContext(ctx, tx, query, arg)
}

type SqlxConnWrapper struct {
	*sqlx.Conn
}

func (conn *SqlxConnWrapper) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return nil, err
	}
	q = conn.Rebind(q)
	return conn.ExecContext(ctx, q, args...)
}

func (conn *SqlxConnWrapper) NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error) {
	q, args, err := sqlx.Named(query, arg)
	if err != nil {
		return nil, err
	}
	q = conn.Rebind(q)
	return conn.QueryxContext(ctx, q, args...)
}

func (conn *SqlxConnWrapper) MustExec(query string, args ...any) sql.Result {
	res, err := conn.ExecContext(context.Background(), query, args...)
	if err != nil {
		panic(err)
	}
	return res
}

func (conn *SqlxConnWrapper) MustExecContext(ctx context.Context, query string, args ...any) sql.Result {
	res, err := conn.ExecContext(ctx, query, args...)
	if err != nil {
		panic(err)
	}
	return res
}
