package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
)

type LazyTxUowRepo struct {
	pgxPool *pgxpool.Pool
	sqlxDB  *sqlx.DB
}

func getStateFromContext(ctx context.Context) *lazyTxUowState {
	state, ok := ctx.Value(lazyTxUowStateCtxKey{}).(*lazyTxUowState)
	if ok && state == nil {
		panic("lazyTxUowState cannot be nil when lazyTxUowStateCtxKey is specify in the context")
	}
	return state
}

func (r *LazyTxUowRepo) Pgx(ctx context.Context) (PgxExecutor, error) {
	if state := getStateFromContext(ctx); state != nil {
		if err := state.initPgxTx(r.pgxPool); err != nil {
			return nil, err
		}
		return state.pgxTx, nil
	}
	return r.pgxPool, nil
}

func (r *LazyTxUowRepo) Sqlx(ctx context.Context) (SqlxExecutor, error) {
	if state := getStateFromContext(ctx); state != nil {
		if err := state.initSqlxTx(r.sqlxDB); err != nil {
			return nil, err
		}
		return &SqlxTx{state.sqlxTx}, nil
	}
	return r.sqlxDB, nil
}

type PgxExecutor interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
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
