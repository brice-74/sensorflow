package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/jmoiron/sqlx"
)

type sqlxTxCtxKey struct{}

func GetSqlxTxFromContext(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(sqlxTxCtxKey{}).(*sqlx.Tx)
	return tx, ok
}

type SqlxTxState struct {
	db     *sqlx.DB
	sqlxTx *sqlx.Tx
	ctx    context.Context
	depth  uint16 // nested transaction depth
}

func NewSqlxTxState(db *sqlx.DB) *SqlxTxState {
	if db == nil {
		panic("*sqlx.DB cannot be nil")
	}
	return &SqlxTxState{
		db:    db,
		depth: 0,
	}
}

func (s *SqlxTxState) Context() context.Context {
	return s.ctx
}

// TransactionFn starts a new transaction block and runs fn inside it.
// Automatically commits/releases if fn succeeds, or rollbacks/reverts if fn fails.
func (s *SqlxTxState) WithTransaction(ctx context.Context, opts *ports.TxUowOptions, fn func(ports.TxUnitOfWork) error) error {
	block, err := s.Transaction(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if revErr := block.Revert(ctx); revErr != nil {
				panic(fmt.Errorf("panic during transaction: %v; rollback error: %w", r, revErr))
			}
			panic(r)
		}
	}()

	if err = fn(block); err != nil {
		if revErr := block.Revert(ctx); revErr != nil {
			err = errors.Join(err, revErr)
		}
		return err
	}

	return block.Finish(ctx)
}

// Transaction starts a new transaction block manually.
// Caller must call Finish() or Revert().
func (s *SqlxTxState) Transaction(ctx context.Context, opts *ports.TxUowOptions) (ports.TxUnitOfWorkFlat, error) {
	if s.depth == 0 {
		var sqlOpts *sql.TxOptions
		if opts != nil {
			sqlOpts = &sql.TxOptions{
				Isolation: toSQLTxIsolation(opts.Isolation),
				ReadOnly:  opts.ReadOnly,
			}
		}
		tx, err := s.db.BeginTxx(ctx, sqlOpts)
		if err != nil {
			return nil, err
		}
		return &SqlxTxState{
			db:     s.db,
			sqlxTx: tx,
			ctx:    context.WithValue(ctx, sqlxTxCtxKey{}, tx),
			depth:  1,
		}, nil
	}

	savepointName := fmt.Sprintf("sp_%d", s.depth)
	if _, err := s.sqlxTx.ExecContext(ctx, "SAVEPOINT "+savepointName); err != nil {
		return nil, err
	}

	return &SqlxTxState{
		db:     s.db,
		sqlxTx: s.sqlxTx,
		ctx:    ctx,
		depth:  s.depth + 1,
	}, nil
}

func toSQLTxIsolation(level ports.UowIsolationLevel) sql.IsolationLevel {
	switch level {
	case ports.IsolationReadCommitted:
		return sql.LevelReadCommitted
	case ports.IsolationRepeatableRead:
		return sql.LevelRepeatableRead
	case ports.IsolationSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}

// Finish finalizes the current transaction block.
func (s *SqlxTxState) Finish(ctx context.Context) error {
	if s.depth == 0 {
		return fmt.Errorf("no active transaction to finish")
	}

	if s.depth == 1 {
		err := s.sqlxTx.Commit()
		s.sqlxTx = nil
		s.depth = 0
		return err
	}

	savepointName := fmt.Sprintf("sp_%d", s.depth-1)
	if _, err := s.sqlxTx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepointName); err != nil {
		return err
	}

	return nil
}

// Revert undoes the current transaction block.
func (s *SqlxTxState) Revert(ctx context.Context) error {
	if s.depth == 0 {
		return fmt.Errorf("no active transaction to revert")
	}

	if s.depth == 1 {
		err := s.sqlxTx.Rollback()
		s.sqlxTx = nil
		s.depth = 0
		return err
	}

	savepointName := fmt.Sprintf("sp_%d", s.depth-1)
	if _, err := s.sqlxTx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepointName); err != nil {
		return err
	}

	return nil
}
