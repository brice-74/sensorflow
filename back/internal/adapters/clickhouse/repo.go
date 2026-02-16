package clickhouse

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

type Repo[T any] interface {
	NewBatch(ctx context.Context) (Batch[T], error)
}

type Batch[T any] interface {
	Append(...T) error
	Flush() error
	Send() error
}

type ConnRegistry interface {
	Get(key uuid.UUID) (*HealthyClient, error)
}

type repo struct {
	registry ConnRegistry
}

func doWithRegistry[T any](ctx context.Context, registry ConnRegistry, dbId uuid.UUID, fn func(context.Context, clickhouse.Conn) (T, error)) (T, error) {
	var zero T
	cli, err := registry.Get(dbId)
	if err != nil {
		return zero, err
	}
	batch, err := fn(ctx, cli.Client)
	if err = cli.HandleError(err); err != nil {
		return zero, err
	}
	return batch, nil
}
