package orchestrators

/*
import (
	"context"
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

var emptyStruct = struct{}{}

type Ingestor[T any] struct {
	logger   log.Logger
	repo     clickhouse.Repo[T]
	backup   BackupIngestor[T]
	syncPool worker.PoolSync[struct{}] // used to limit concurrent call by db
	rowsBuf  xsync.BatchSlice[T]

	stopCh       chan struct{}
	notifyCh     chan struct{}
	backupStopCh chan struct{}
	batch        clickhouse.Batch[T]

	tick            time.Duration
	maxFlushRows    int
	FlushDelay      time.Duration
	maxRotationRows int
	RotationDelay   time.Duration
	lastFlush       atomic.Int64 // UnixNano
	lastRotation    atomic.Int64 // UnixNano

	backupTriggerThreshold int
	backupBatchSize        int
	backupDrainDelay       time.Duration

	graceActive atomic.Bool
	degraded    atomic.Bool
	isStopped   atomic.Bool
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
	i.stopBackupTimer()

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

func (i *Ingestor[T]) State() IngestorState {
	if i.backupProcessing.Load() {
		return IngestorDegraded
	}
	if i.backupTimerActive.Load() {
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
		case <-ticker.C:
			if i.degraded.Load() {

			} else {
				i.checkTimeOnly()
			}
		case <-i.notifyCh:
			if i.degraded.Load() {

			} else {
				i.checkCountOnly()
				i.backupExcessIfNeeded()
			}
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
		now.Sub(time.Unix(0, i.lastFlush.Load())) >= i.FlushDelay {
		i.flush()
	}

	if i.batch != nil && i.batch.Rows() > 0 &&
		now.Sub(time.Unix(0, i.lastRotation.Load())) >= i.RotationDelay {
		i.rotate()
	}
}

func (i *Ingestor[T]) flush() {
	rows := i.rowsBuf.PeekBatch(i.maxFlushRows)
	if len(rows) == 0 {
		return
	}

	_, err := i.syncPool.Submit(func() (struct{}, error) {
		if !i.repo.Client().IsAlive() {
			return emptyStruct, errors.WrapMsg("db down")
		}

		var err error
		if i.batch == nil {
			i.batch, err = i.repo.NewBatch(context.Background())
			if err != nil {
				return emptyStruct, err
			}
		}

		if err := i.batch.Append(rows...); err != nil {
			return emptyStruct, err
		}

		if err := i.batch.Flush(); err != nil {
			return emptyStruct, err
		}

		return emptyStruct, nil
	})

	if err != nil {
		i.startGraceTimer()
		i.logger.Error(errors.WrapErr(err))
		return
	}

	i.rowsBuf.CommitAndRelease(len(rows))
	i.lastFlush.Store(time.Now().UnixNano())

	i.stopBackupTimer()
}

func (i *Ingestor[T]) rotate() {

}

func (i *Ingestor[T]) startGraceTimer() {
	if i.graceActive.Load() || i.degraded.Load() {
		return
	}

	i.graceActive.Store(true)
	i.backupStopCh = make(chan struct{})

	go func() {
		timer := time.NewTimer(i.backupDrainDelay)
		defer timer.Stop()

		select {
		case <-timer.C:
			i.degraded.Store(true)
			i.graceActive.Store(false)
		case <-i.backupStopCh:
			i.graceActive.Store(false)
			return
		}
	}()
}

func (i *Ingestor[T]) backupDrainLoop() {
	i.logger.Warn(errors.WrapMsg("Entering degraded backup mode"))

	for {
		select {
		case <-i.backupStopCh:
			return
		default:
		}

		if !i.degraded.Load() || i.isStopped.Load() {
			return
		}

		if i.rowsBuf.Len() == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		rows := i.rowsBuf.PeekBatch(i.backupBatchSize)
		if len(rows) == 0 {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if err := i.backup.Store(rows); err != nil {
			i.logger.Error(errors.WrapErr(err))
			time.Sleep(500 * time.Millisecond)
			continue
		}

		i.rowsBuf.CommitAndRelease(len(rows))
	}
}

func (i *Ingestor[T]) backupExcessIfNeeded() {
	for i.rowsBuf.Len() > i.backupTriggerThreshold {
		excess := i.rowsBuf.Len() - i.backupTriggerThreshold
		batchSize := min(excess, i.backupBatchSize)

		rows := i.rowsBuf.PeekBatch(batchSize)
		if len(rows) == 0 {
			break
		}

		if err := i.backup.Store(rows); err != nil {
			i.logger.Error(errors.WrapErr(err))
			break
		}

		i.rowsBuf.CommitAndRelease(len(rows))
		i.logger.Warn(errors.WrapMsg("One-off backup"))
	}
}

func (i *Ingestor[T]) stopBackupTimer() {
	if i.backupStopCh != nil {
		close(i.backupStopCh)
		i.backupStopCh = nil
	}

	i.degraded.Store(false)
	i.graceActive.Store(false)
}
*/
