package orchestrators

import (
	"sync/atomic"
	"time"

	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/xsync"
)

type IngestStatus uint8

const (
	INGEST_STATUS_UNSPECIFIED IngestStatus = iota
	INGEST_STATUS_OK
	INGEST_STATUS_PENDING
	INGEST_STATUS_RETRYING_MAIN
	INGEST_STATUS_RETRYING_LEVEL
	INGEST_STATUS_LEVEL_FALLBACK
	INGEST_STATUS_ERROR
	INGEST_STATUS_STOPPED
)

type IngestionLevel[T any] interface {
	Name() string
	Write(rows []T) error
}

type Ingestor[T any] struct {
	logger  log.Logger
	rowsBuf xsync.BatchSlice[T]

	levels       []IngestionLevel[T]
	currentLevel int

	tick           time.Duration
	flushBatchSize int
	flushDelay     time.Duration
	lastFlushAt    atomic.Int64

	retryDelay              time.Duration
	nextRetryAt             atomic.Int64
	maxBufferBeforeFallback int

	levelProbeInterval time.Duration
	lastProbeAt        atomic.Int64
	lastFlushFailedAt  atomic.Int64
	stopped            atomic.Bool

	stopCh   chan struct{}
	notifyCh chan struct{}
}

func (i *Ingestor[T]) Start() {
	i.stopCh = make(chan struct{})
	i.notifyCh = make(chan struct{}, 1)
	i.stopped.Store(false)

	now := time.Now()
	i.lastFlushAt.Store(now.UnixNano())
	i.lastProbeAt.Store(now.UnixNano())

	go i.loop()
}

func (i *Ingestor[T]) Stop() {
	i.stopped.Store(true)
	close(i.stopCh)
	close(i.notifyCh)
}

func (i *Ingestor[T]) Submit(rows ...T) {
	i.rowsBuf.Append(rows...)
	select {
	case i.notifyCh <- struct{}{}:
	default:
	}
}

func (i *Ingestor[T]) State() IngestStatus {
	now := time.Now()
	rows := i.rowsBuf.Len()

	if i.stopped.Load() {
		return INGEST_STATUS_STOPPED
	}

	if failedAt := i.lastFlushFailedAt.Load(); failedAt > 0 && failedAt > i.lastFlushAt.Load() {
		return INGEST_STATUS_ERROR
	}

	if now.UnixNano() < i.nextRetryAt.Load() {
		if i.currentLevel == 0 {
			return INGEST_STATUS_RETRYING_MAIN
		}
		return INGEST_STATUS_RETRYING_LEVEL
	}

	if i.currentLevel > 0 {
		return INGEST_STATUS_LEVEL_FALLBACK
	}

	if i.maxBufferBeforeFallback > 0 {
		pressure := float64(rows) / float64(i.maxBufferBeforeFallback)
		if pressure >= 0.5 {
			return INGEST_STATUS_PENDING
		}
	}

	return INGEST_STATUS_OK
}

func (i *Ingestor[T]) loop() {
	ticker := time.NewTicker(i.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			i.process()
		case <-i.notifyCh:
			i.process()
		case <-i.stopCh:
			return
		}
	}
}

// process decides if we need to flush based on count or timer
func (i *Ingestor[T]) process() {
	lenBuf := i.rowsBuf.Len()
	if lenBuf == 0 {
		return
	}

	now := time.Now()
	if lenBuf >= i.flushBatchSize || now.Sub(time.Unix(0, i.lastFlushAt.Load())) >= i.flushDelay {
		i.flushWithPolicy(now)
	}
}

// flushWithPolicy handles writing rows to levels with retry/fallback
func (i *Ingestor[T]) flushWithPolicy(now time.Time) {
	rows := i.rowsBuf.PeekBatch(i.flushBatchSize)
	if len(rows) == 0 {
		return
	}

	i.maybePromote(now)

	nowNano := now.UnixNano()

	for lvl := i.currentLevel; lvl < len(i.levels); lvl++ {
		// respect retry delay for current level
		if lvl == i.currentLevel && nowNano < i.nextRetryAt.Load() {
			return
		}

		err := i.levels[lvl].Write(rows)
		if err == nil {
			i.currentLevel = lvl
			i.rowsBuf.CommitAndRelease(len(rows))
			i.lastFlushAt.Store(nowNano)
			i.nextRetryAt.Store(0)
			return
		}

		i.logger.Warn(errors.Wrapf(err, "ingestion level %s failed", i.levels[lvl].Name()))

		// manage fallback only for current level
		if lvl == i.currentLevel {
			i.nextRetryAt.Store(now.Add(i.retryDelay).UnixNano())

			// fallback to next level if buffer too big
			if i.rowsBuf.Len() >= i.maxBufferBeforeFallback && lvl+1 < len(i.levels) {
				i.logger.Warn(errors.WrapMsg("falling back to next ingestion level"))
				i.currentLevel++
				return
			}

			return
		}
	}

	i.lastFlushFailedAt.Store(nowNano)
	i.logger.Error(errors.WrapMsg("all ingestion levels failed"))
}

// maybePromote attempts to move back to higher levels if enough time has passed
func (i *Ingestor[T]) maybePromote(now time.Time) {
	if i.currentLevel == 0 {
		return
	}

	lastProbe := time.Unix(0, i.lastProbeAt.Load())
	if now.Sub(lastProbe) < i.levelProbeInterval {
		return
	}

	i.lastProbeAt.Store(now.UnixNano())
	i.currentLevel = 0
	i.nextRetryAt.Store(0)
}
