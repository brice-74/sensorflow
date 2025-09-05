package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
)

type lazyTxUowStateCtxKey struct{}

type lazyTxUowState struct {
	pgxTx  pgx.Tx
	sqlxTx *sqlx.Tx
	ctx    context.Context
	mu     sync.Mutex
	// current depth for savepoints
	depth uint16
	// used for sqlx
	savepointCreated bool
}

func (s *lazyTxUowState) Context() context.Context {
	return s.ctx
}

func (s *lazyTxUowState) TransactionFn(ctx context.Context, fn func(ports.TxUnitOfWork) error) error {
	tx, err := s.Transaction(ctx)
	if err != nil {
		return err
	}

	defer tx.Revert(ctx)
	if err := fn(tx); err != nil {
		return err
	}

	return tx.Finish(ctx)
}

func (s *lazyTxUowState) Transaction(ctx context.Context) (ports.TxUnitOfWorkFlat, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	nested := &lazyTxUowState{
		pgxTx:  s.pgxTx,
		sqlxTx: s.sqlxTx,
		depth:  s.depth + 1,
		ctx:    ctx,
	}

	return nested, nil
}

// Finish commits or releases transactions/savepoints
func (s *lazyTxUowState) Finish(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	// pgx already handle nested transactions with savepoints
	if s.pgxTx != nil {
		if e := s.pgxTx.Commit(ctx); e != nil {
			err = e
		}
		s.pgxTx = nil
	}

	if s.sqlxTx != nil {
		if s.depth == 1 {
			if e := s.sqlxTx.Commit(); e != nil && err == nil {
				err = e
			}
			s.sqlxTx = nil
		} else {
			sp := s.savepointName()
			if e := s.releaseSavepoint(ctx, sp); e != nil && err == nil {
				err = e
			}
		}
	}

	return err
}

// Revert rolls back transactions/savepoints
func (s *lazyTxUowState) Revert(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	if s.pgxTx != nil {
		if e := s.pgxTx.Rollback(ctx); e != nil {
			err = e
		}
		s.pgxTx = nil
	}

	if s.sqlxTx != nil {
		if s.depth == 1 {
			if e := s.sqlxTx.Rollback(); e != nil && err == nil {
				err = e
			}
			s.sqlxTx = nil
		} else {
			sp := s.savepointName()
			if e := s.rollbackToSavepoint(ctx, sp); e != nil && err == nil {
				err = e
			}
		}
	}

	return err
}

func (s *lazyTxUowState) ensureSavepoint(ctx context.Context) error {
	if s.sqlxTx != nil && s.depth > 1 && !s.savepointCreated {
		sp := s.savepointName()
		if _, err := s.sqlxTx.ExecContext(ctx, fmt.Sprintf("SAVEPOINT %s", sp)); err != nil {
			return err
		}
		s.savepointCreated = true
	}
	return nil
}

func (s *lazyTxUowState) savepointName() string {
	return fmt.Sprintf("sp_%d", s.depth)
}

func (s *lazyTxUowState) rollbackToSavepoint(ctx context.Context, sp string) error {
	_, err := s.sqlxTx.ExecContext(ctx, fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", sp))
	return err
}

func (s *lazyTxUowState) releaseSavepoint(ctx context.Context, sp string) error {
	_, err := s.sqlxTx.ExecContext(ctx, fmt.Sprintf("RELEASE SAVEPOINT %s", sp))
	return err
}

// -------------------- Lazy init helpers --------------------

func (s *lazyTxUowState) initPgxTx(pool *pgxpool.Pool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pgxTx == nil {
		tx, err := pool.Begin(s.ctx)
		if err != nil {
			return err
		}
		s.pgxTx = tx
	}
	return nil
}

func (s *lazyTxUowState) initSqlxTx(db *sqlx.DB) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.sqlxTx == nil {
		tx, err := db.Beginx()
		if err != nil {
			return err
		}
		s.sqlxTx = tx
	}
	return nil
}
