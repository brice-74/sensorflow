package postgres

import "context"

type SqlxManagerImpl struct {
	*SqlxTxManager
}

var _ SqlxManager = (*SqlxManagerImpl)(nil)

func (m *SqlxManagerImpl) WithConnection(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithConnection(ctx, m.db, fn)
}
func (m *SqlxManagerImpl) Connection(ctx context.Context) (func() error, context.Context, error) {
	return Connection(ctx, m.db)
}
