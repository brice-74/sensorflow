package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type sqlxRepo struct {
	DB *sqlx.DB
}

func NewSqlxRepo(db *sqlx.DB) *sqlxRepo {
	return &sqlxRepo{DB: db}
}

func (r *sqlxRepo) ExecutorFromCtx(ctx context.Context) SqlxExecutor {
	if tx, ok := GetSqlxTx(ctx); ok && tx != nil {
		return &SqlxTxWrapper{tx}
	}
	if conn, ok := GetSqlxConn(ctx); ok && conn != nil {
		return &SqlxConnWrapper{conn}
	}
	return r.DB
}

func (r *sqlxRepo) TxFromCtx(ctx context.Context) (*SqlxTxWrapper, bool) {
	tx, ok := GetSqlxTx(ctx)
	return &SqlxTxWrapper{tx}, ok
}

func (r *sqlxRepo) ConnFromCtx(ctx context.Context) (*SqlxConnWrapper, bool) {
	conn, ok := GetSqlxConn(ctx)
	return &SqlxConnWrapper{conn}, ok
}
