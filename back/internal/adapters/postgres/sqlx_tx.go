package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/jmoiron/sqlx"
)

type sqlxTxCtxKeytype struct{}

var sqlxTxCtxKey = sqlxTxCtxKeytype{}

func GetSqlxTx(ctx context.Context) (*sqlx.Tx, bool) {
	tx, ok := ctx.Value(sqlxTxCtxKey).(*sqlx.Tx)
	return tx, ok
}

type SqlxTxManager struct {
	db    *sqlx.DB
	tx    *sqlx.Tx
	depth int
}

var _ SqlxTx = (*SqlxTxManager)(nil)

func NewSqlxTxManager(db *sqlx.DB) *SqlxTxManager {
	return &SqlxTxManager{db: db}
}

// WithTransaction runs fn inside a transaction or savepoint automatically.
func (t *SqlxTxManager) WithTransaction(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	sub, err := t.Begin(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			if subErr := sub.Rollback(ctx); subErr != nil {
				panic(fmt.Sprintf("panic during transaction: %v; rollback error: %v", r, subErr))
			}

			panic(r)
		}
	}()

	if err := fn(sub.Context()); err != nil {
		if subErr := sub.Rollback(ctx); subErr != nil {
			err = errors.Join(err, subErr)
		}

		return err
	}

	return sub.Commit(ctx)
}

func (t *SqlxTxManager) Begin(ctx context.Context, opts *sql.TxOptions) (SqlxTxFlat, error) {
	return t.BeginC(ctx, opts)
}

// begin starts a transaction or nested savepoint.
func (t *SqlxTxManager) BeginC(ctx context.Context, opts *sql.TxOptions) (*SqlxTxManager, error) {
	if t.depth == 0 {
		tx, err := t.db.BeginTxx(ctx, opts)
		if err != nil {
			return nil, errors.Wrap(err, "begin transaction")
		}
		return &SqlxTxManager{
			db:    t.db,
			tx:    tx,
			depth: 1,
		}, nil
	}

	savepoint := fmt.Sprintf("sp_%d", t.depth)
	if _, err := t.tx.ExecContext(ctx, "SAVEPOINT "+savepoint); err != nil {
		return nil, errors.Wrapf(err, "create savepoint %s", savepoint)
	}
	return &SqlxTxManager{
		db:    t.db,
		tx:    t.tx,
		depth: t.depth + 1,
	}, nil
}

func (t *SqlxTxManager) Tx() *sqlx.Tx {
	return t.tx
}

func (t *SqlxTxManager) Context() context.Context {
	return context.WithValue(context.Background(), sqlxTxCtxKey, t.tx)
}

func (t *SqlxTxManager) Commit(ctx context.Context) error {
	if t.depth == 1 {
		return errors.Wrap(t.tx.Commit(), "commit transaction")
	}
	savepoint := fmt.Sprintf("sp_%d", t.depth-1)
	_, err := t.tx.ExecContext(ctx, "RELEASE SAVEPOINT "+savepoint)
	return errors.Wrap(err, "release savepoint")
}

func (t *SqlxTxManager) Rollback(ctx context.Context) error {
	if t.depth == 1 {
		return errors.Wrap(t.tx.Rollback(), "rollback transaction")
	}
	savepoint := fmt.Sprintf("sp_%d", t.depth-1)
	_, err := t.tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT "+savepoint)
	return errors.Wrap(err, "rollback savepoint")
}
