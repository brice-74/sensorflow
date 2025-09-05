package ports

import "context"

type TxUnitOfWorkFlat interface {
	TxUnitOfWork
	// Finish finalizes the current unit of work (commit or release).
	Finish(ctx context.Context) error
	// Revert undoes the current unit of work (rollback or revert to previous state).
	Revert(ctx context.Context) error
}

type TxUnitOfWork interface {
	// TransactionFn executes the given function within a new unit of work.
	// Automatically finalizes or reverts based on function result.
	TransactionFn(ctx context.Context, fn func(TxUnitOfWork) error) error
	// Transaction starts a new unit of work manually.
	// Caller is responsible for calling Finish or Revert.
	Transaction(ctx context.Context) (TxUnitOfWorkFlat, error)
	// Context returns the current context associated with this unit of work.
	// It can be used to carry metadata or propagate state to dependent operations.
	Context() context.Context
}
