package ingestor

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
)

type ClickhouseLevel[T any] struct {
	repo   clickhouse.Repo[T]
	logger log.Logger

	batch clickhouse.Batch[T]

	flushCount         int
	maxFlushBeforeSend int

	lastSend         time.Time
	maxBatchLifetime time.Duration
}

var _ IngestionLevel[any] = (*ClickhouseLevel[any])(nil)

func NewClickhouseLevel[T any](repo clickhouse.Repo[T], logger log.Logger, maxFlushBeforeSend int, maxBatchLifetime time.Duration) *ClickhouseLevel[T] {
	return &ClickhouseLevel[T]{
		repo:               repo,
		logger:             logger,
		maxFlushBeforeSend: maxFlushBeforeSend,
		maxBatchLifetime:   maxBatchLifetime,
	}
}

func (l *ClickhouseLevel[T]) Name() string {
	return "clickhouse"
}

func (l *ClickhouseLevel[T]) Write(ctx context.Context, rows []T) error {
	if len(rows) == 0 {
		return nil
	}

	if !l.repo.Client().IsAlive() {
		return errors.WrapMsg("db down")
	}

	var err error

	if l.batch == nil {
		l.batch, err = l.repo.NewBatch(ctx)
		if err != nil {
			return errors.Wrap(err, "create batch failed")
		}
		l.lastSend = time.Now()
	}

	if err := l.batch.Append(rows...); err != nil {
		l.closeBatch()
		return errors.Wrap(err, "append failed")
	}

	if l.shouldRotate() {
		return l.sendAndClose()
	}

	if err := l.batch.Flush(); err != nil {
		l.closeBatch()
		return errors.Wrap(err, "flush failed")
	}

	l.flushCount++

	return nil
}

func (l *ClickhouseLevel[T]) shouldRotate() bool {
	if l.maxFlushBeforeSend > 0 && l.flushCount >= l.maxFlushBeforeSend {
		return true
	}

	if l.maxBatchLifetime > 0 && time.Since(l.lastSend) >= l.maxBatchLifetime {
		return true
	}

	return false
}

func (l *ClickhouseLevel[T]) closeBatch() {
	if l.batch != nil {
		err := l.batch.Close()
		if err != nil {
			l.logger.Warn(errors.WrapErr(err))
		}
	}

	l.batch = nil
	l.flushCount = 0
}

func (l *ClickhouseLevel[T]) sendAndClose() error {
	if l.batch == nil || l.batch.IsSent() {
		return nil
	}

	if err := l.batch.Send(); err != nil {
		l.closeBatch()
		return errors.Wrap(err, "send failed")
	}

	l.closeBatch()
	l.lastSend = time.Now()

	return nil
}
