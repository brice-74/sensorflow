package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type SqlxRepo struct {
	DB *sqlx.DB
}

func (r *SqlxRepo) Executor(ctx context.Context) SqlxExecutor {
	if tx, ok := GetSqlxTx(ctx); ok && tx != nil {
		return &SqlxTx{tx}
	}
	if conn, ok := GetSqlxConn(ctx); ok && conn != nil {
		return &SqlxConn{conn}
	}
	return r.DB
}

func (r *SqlxRepo) Tx(ctx context.Context) (*SqlxTx, bool) {
	tx, ok := GetSqlxTx(ctx)
	return &SqlxTx{tx}, ok
}

func (r *SqlxRepo) Conn(ctx context.Context) (*SqlxConn, bool) {
	conn, ok := GetSqlxConn(ctx)
	return &SqlxConn{conn}, ok
}

type SqlxExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	MustExecContext(ctx context.Context, query string, args ...any) sql.Result
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	NamedQueryContext(ctx context.Context, query string, arg any) (*sqlx.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	Rebind(query string) string
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

type PgxExecutor interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
}
