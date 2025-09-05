package postgres

/* import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
)

type stateCtxKey struct{}

func getState(ctx context.Context) *uowState {
	state, ok := ctx.Value(stateCtxKey{}).(*uowState)
	if !ok {
		panic("UnitOfWork state not found in context")
	}
	return state
}

type uowState struct {
	useTx, useConn bool
	sqlxTx         *sqlx.Tx
	pgxTx          pgx.Tx
	sqlxConn       *sqlx.Conn
	pgxConn        *pgxpool.Conn
}

// runOps exécute deux opérations (sqlx + pgx) et retourne une erreur jointe si besoin
func (s *uowState) runOps(sqlxOp func() error, pgxOp func() error) error {
	var err1, err2 error

	if s.sqlxTx != nil && sqlxOp != nil {
		if err := sqlxOp(); err != nil && err != sql.ErrTxDone {
			err1 = err
		}
	}
	if s.pgxTx != nil && pgxOp != nil {
		if err := pgxOp(); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			err2 = err
		}
	}

	if err1 != nil && err2 != nil {
		return errors.Join(err1, err2)
	}
	if err1 != nil {
		return err1
	}
	return err2
}

func (s *uowState) commit(ctx context.Context) error {
	return s.runOps(
		func() error { return s.sqlxTx.Commit() },
		func() error { return s.pgxTx.Commit(ctx) },
	)
}

func (s *uowState) rollback(ctx context.Context) error {
	return s.runOps(
		func() error { return s.sqlxTx.Rollback() },
		func() error { return s.pgxTx.Rollback(ctx) },
	)
}

func (s *uowState) release() error {
	var err error
	if s.sqlxConn != nil {
		if cerr := s.sqlxConn.Close(); cerr != nil {
			err = errors.Join(err, fmt.Errorf("sqlx conn close: %w", cerr))
		}
	}
	if s.pgxConn != nil {
		s.pgxConn.Release()
	}
	return err
}

// --- Impl ---

type uowImpl struct {
	sqlxDB  *sqlx.DB
	pgxPool *pgxpool.Pool
}

func (u *uowImpl) useState(ctx context.Context, state *uowState, fn func(ctx context.Context) error, cleanup func(context.Context) error) error {
	ctx = context.WithValue(ctx, stateCtxKey{}, state)
	if err := fn(ctx); err != nil {
		if cerr := cleanup(ctx); cerr != nil {
			return errors.Join(err, cerr)
		}
		return err
	}
	return cleanup(ctx)
}

func (u *uowImpl) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.useState(ctx, &uowState{useTx: true}, fn, func(ctx context.Context) error {
		return getState(ctx).commit(ctx)
	})
}

func (u *uowImpl) WithConn(ctx context.Context, fn func(ctx context.Context) error) error {
	return u.useState(ctx, &uowState{useConn: true}, fn, func(ctx context.Context) error {
		return getState(ctx).release()
	})
}

// --- Low-level API ---
func (u *uowImpl) BeginTx(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, stateCtxKey{}, &uowState{useTx: true}), nil
}

func (u *uowImpl) CommitTx(ctx context.Context) error {
	return getState(ctx).commit(ctx)
}

func (u *uowImpl) RollbackTx(ctx context.Context) error {
	return getState(ctx).rollback(ctx)
}

func (u *uowImpl) AcquireConn(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, stateCtxKey{}, &uowState{useConn: true}), nil
}

func (u *uowImpl) ReleaseConn(ctx context.Context) error {
	return getState(ctx).release()
}

type UowRepo struct {
}

type uowContext struct {
	pgxPool *pgxpool.Pool
	sqlxDB  *sqlx.DB
}

func (d *uowContext) Sqlx(ctx context.Context) SQLXExecutor {
	if tx := ctx.Value(ctxKeySqlxTx); tx != nil {
		return tx.(*SqlxTx)
	}
	if conn := ctx.Value(ctxKeySqlxConn); conn != nil {
		return conn.(*SqlxConn)
	}
	return d.sqlxDB
}

func (d *uowContext) Pgx(ctx context.Context) PGXExecutor {
	if tx := ctx.Value(ctxKeyPgxTx); tx != nil {
		return tx.(pgx.Tx)
	}
	if conn := ctx.Value(ctxKeyPgxConn); conn != nil {
		return conn.(*pgxpool.Conn)
	}
	return d.pgxPool
}

type UnitOfWork interface {
	UseTxFn(ctx context.Context, fn func(ctx context.Context) error) error
	UseConnFn(ctx context.Context, fn func(ctx context.Context) error) error

	UseTx(ctx context.Context) (context.Context, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	UseConn(ctx context.Context) (context.Context, error)
	ReleaseConn(ctx context.Context) error
}

type UowContext interface {
	Sqlx(ctx context.Context) SqlxExecutor
	Pgx(ctx context.Context) PgxExecutor
}

type PgxExecutor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
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
*/
