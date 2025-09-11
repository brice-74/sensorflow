package app

import (
	"github.com/brice-74/sensorflow/internal/adapters/postgres"
	"github.com/brice-74/sensorflow/internal/core/ports"
	"github.com/jmoiron/sqlx"
)

type UnitOfWorks struct {
	TransactionUnitOfWork ports.TxUnitOfWork
	ConnectionUnitOfWork  ports.ConnUnitOfWork
}

func NewUnitOfWorks(sqlxDB *sqlx.DB) *UnitOfWorks {
	uows := UnitOfWorks{
		TransactionUnitOfWork: postgres.NewSqlxTxState(sqlxDB),
		ConnectionUnitOfWork:  postgres.NewSqlxConnState(sqlxDB),
	}

	return &uows
}
