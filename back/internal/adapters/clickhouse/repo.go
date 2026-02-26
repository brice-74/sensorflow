package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

type Repo[T any] interface {
	NewBatch(ctx context.Context) (Batch[T], error)
	Client() *HealthyClient
}

type DynamicRepo[T any] interface {
	NewBatch(ctx context.Context, dbID uuid.UUID) (Batch[T], error)
	Registry() ConnRegistry
}

type Batch[T any] interface {
	Append(...T) error
	Flush() error
	Send() error
	Rows() int
}

type ConnRegistry interface {
	Get(key uuid.UUID) (*HealthyClient, error)
}

type repo struct {
	client *HealthyClient
}

type dynamicRepo struct {
	registry ConnRegistry
}

func (r *dynamicRepo) Registry() ConnRegistry {
	return r.registry
}

func doWithRegistry[T any](ctx context.Context, registry ConnRegistry, dbId uuid.UUID, fn func(context.Context, clickhouse.Conn) (T, error)) (T, error) {
	var zero T
	c, err := registry.Get(dbId)
	if err != nil {
		return zero, err
	}
	batch, err := fn(ctx, c.Client)
	if err = c.HandleError(err); err != nil {
		return zero, err
	}
	return batch, nil
}
