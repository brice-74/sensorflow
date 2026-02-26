package orchestrators

import (
	"sync/atomic"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/worker"
	"github.com/brice-74/sensorflow/pkg/xsync"
)

type BackupIngestor[T any] interface {
	Store(rows []T) error
}

type IngestorState uint32

const (
	IngestorOff IngestorState = iota
	IngestorOn
	IngestorThrottled
	IngestorDegraded
)

type Ingestor[T any] struct {
	logger   log.Logger
	repo     clickhouse.Repo[T]
	backup   BackupIngestor[T]
	taskPool worker.ResultTaskSession[any]
	rowsBuf  xsync.BatchSlice[T]

	stopCh   chan struct{}
	notifyCh chan struct{}
	batch    clickhouse.Batch[T]

	tick            time.Duration
	maxFlushRows    int
	maxFlushWait    time.Duration
	maxRotationRows int
	maxRotationWait time.Duration
	lastFlush       atomic.Int64 // UnixNano
	lastRotation    atomic.Int64 // UnixNano

	// State flags
	externalDown   atomic.Bool
	queueThrottled atomic.Bool
	isStopped      atomic.Bool
}

func (i *Ingestor[T]) Start() {
	i.stopCh = make(chan struct{})
	i.notifyCh = make(chan struct{}, 1)

	now := time.Now().UnixNano()
	i.lastFlush.Store(now)
	i.lastRotation.Store(now)

	i.isStopped.Store(false)

	go i.loop()
}

func (i *Ingestor[T]) Stop() {
	close(i.stopCh)
	close(i.notifyCh)

	i.isStopped.Store(true)
}

// Submit adds rows to the ingestor's buffer by sending non-blocking notifications to the internal loop.
// It panics if the ingestor isn't started or stopped.
func (i *Ingestor[T]) Submit(rows ...T) {
	i.rowsBuf.Append(rows...)

	select {
	case i.notifyCh <- struct{}{}:
	default:
	}
}

func (i *Ingestor[T]) currentState() IngestorState {
	if i.externalDown.Load() {
		return IngestorDegraded
	}

	if i.queueThrottled.Load() {
		return IngestorThrottled
	}

	if i.isStopped.Load() {
		return IngestorOff
	}

	return IngestorOn
}

func (i *Ingestor[T]) loop() {
	ticker := time.NewTicker(i.tick)
	defer ticker.Stop()

	for {
		select {
		case isAlive := <-i.repo.Client().IsAliveCh():
			i.externalDown.Store(!isAlive)
		case <-ticker.C:
			i.checkTimeOnly()
		case <-i.notifyCh:
			i.checkCountOnly()
		case <-i.stopCh:
			return
		}
	}
}

func (i *Ingestor[T]) checkCountOnly() {
	if i.batch != nil && i.batch.Rows() >= i.maxRotationRows {
		i.rotate()
	}

	if i.rowsBuf.Len() >= i.maxFlushRows {
		i.flush()
	}
}

func (i *Ingestor[T]) checkTimeOnly() {
	now := time.Now()

	if i.rowsBuf.Len() > 0 &&
		now.Sub(time.Unix(0, i.lastFlush.Load())) >= i.maxFlushWait {
		i.flush()
	}

	if i.batch != nil && i.batch.Rows() > 0 &&
		now.Sub(time.Unix(0, i.lastRotation.Load())) >= i.maxRotationWait {
		i.rotate()
	}
}

func (i *Ingestor[T]) flush() {
	rows := i.rowsBuf.TakeBatch(i.maxFlushRows)
	defer func() {
		i.rowsBuf.Release(rows)
		i.lastFlush.Store(time.Now().UnixNano())
	}()

	err := i.taskPool.Submit(func() any {
		if !i.repo.Client().IsAlive() {
			i.doBackup(rows)
			return nil
		}

		var err error
		if i.batch == nil {
			i.batch, err = i.repo.NewBatch(nil) // TODO: pass context
			if err != nil {
				i.logger.Error(errors.WrapErr(err))
				i.doBackup(rows)
				return nil
			}
		}

		err = i.batch.Append(rows...)
		if err != nil {
			i.logger.Error(errors.WrapErr(err))
			i.doBackup(rows)
			return nil
		}

		err = i.batch.Flush()
		if err != nil {
			i.logger.Error(errors.WrapErr(err))
			i.doBackup(rows)
			return nil
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, worker.ErrQueueFull) {
			i.queueThrottled.Store(true)
			return
		}
		i.logger.Error(errors.WrapErr(err))
	}
}

func (i *Ingestor[T]) rotate() {

}

func (i *Ingestor[T]) doBackup(rows []T) {
	if err := i.backup.Store(rows); err != nil {
		i.logger.Error(errors.WrapErr(err))
	}
}

func (i *Ingestor[T]) State() IngestorState {
	return i.currentState()
}
