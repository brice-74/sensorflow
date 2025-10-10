package ports

import "context"

// UowIsolationLevel represents the isolation level for a unit of work.
// Its meaning and support depend on the underlying implementation.
type UowIsolationLevel int

const (
	// IsolationDefault lets the implementation decide the default isolation.
	IsolationDefault UowIsolationLevel = iota
	// IsolationReadCommitted guarantees reads only see committed data.
	IsolationReadCommitted
	// IsolationRepeatableRead guarantees all reads in the unit see a consistent snapshot.
	IsolationRepeatableRead
	// IsolationSerializable provides full serializability of operations.
	IsolationSerializable
)

// TxUowOptions defines optional parameters for configuring a unit of work.
// These are only applied to the root unit; nested units inherit the parent's configuration.
type TxUowOptions struct {
	// ReadOnly indicates whether the unit of work should be read-only.
	ReadOnly bool
	// Isolation specifies the isolation level; meaning depends on the implementation.
	Isolation UowIsolationLevel
}

// TxUnitOfWorkFlat extends TxUnitOfWork with explicit control over the
// current unit-of-work block. Finish() finalizes the block, and Revert()
// undoes it. See the comment on TxUnitOfWork for context propagation details.
type TxUnitOfWorkFlat interface {
	TxUnitOfWork
	// Finish finalizes the current unit of work block.
	// - If this is the root unit, it commits/persists the changes.
	// - If nested, it validates/releases only the current block.
	Finish(ctx context.Context) error
	// Revert undoes the current unit of work block.
	// - If this is the root unit, it fully rolls back/reverts all changes.
	// - If nested, it reverts to the state before this block began.
	Revert(ctx context.Context) error
}

// TxUnitOfWork defines a generic unit-of-work abstraction for executing
// code inside nested blocks. Context() exposes the current context, which
// must be propagated to sub-blocks to share the same underlying state.
// Concurrent operations on the same unit-of-work instance are not supported,
// to prevent inconsistent state or race conditions.
// For parallel work, create separate unit-of-work instances.
// Options can be provided when starting a block; how these options are
// interpreted depends on the implementation. In SQL-based implementations,
// certain options (e.g., isolation level) typically only apply to the root
// transaction.
//
// Example:
//
//	root.WithTransaction(root.Context(), &TxUowOptions{}, func(uow TxUnitOfWork) error {
//	    return uow.WithTransaction(uow.Context(), nil, func(inner TxUnitOfWork) error {
//	        doSomething(inner.Context())
//	        return nil
//	    })
//	})
type TxUnitOfWork interface {
	// WithTransaction executes the given function inside a new unit of work block.
	// Automatically calls Finish if fn succeeds, or Revert if it returns an error.
	// Nested calls create sub-blocks that can be finalized or reverted independently.
	WithTransaction(ctx context.Context, opts *TxUowOptions, fn func(TxUnitOfWork) error) error
	// Transaction starts a new unit of work block manually.
	// The caller must explicitly call Finish() or Revert().
	// Nested calls should inherit configuration from the parent block.
	Transaction(ctx context.Context, opts *TxUowOptions) (TxUnitOfWorkFlat, error)
	// Context returns the context associated with the current unit of work.
	// This must be passed to any sub-blocks or dependent operations to share the same state.
	Context() context.Context
}

// ConnUnitOfWork provides controlled access to a shared connection or resource.
// It can be used independently or in combination with a transaction.
type ConnUnitOfWork interface {
	// WithConnection executes fn using the current connection or unit of work.
	// Automatically handles opening and releasing the connection after fn completes.
	WithConnection(ctx context.Context, fn func(ctx context.Context) error) error
	// Connection returns the current connection, along with a function to release it.
	// The returned ctxOut may contain additional values or state tied to the connection.
	Connection(ctx context.Context) (releaseFn func() error, ctxOut context.Context, err error)
}
