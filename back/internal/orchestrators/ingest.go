package orchestrators

import (
	"sync/atomic"
	"time"

	"github.com/brice-74/sensorflow/internal/log"
	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/brice-74/sensorflow/pkg/xsync"
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

	// flush config
	tick         time.Duration
	maxFlushRows int
	flushDelay   time.Duration
	lastFlush    atomic.Int64

	// retry / fallback policy
	retryDelay              time.Duration
	nextRetryAt             atomic.Int64
	maxBufferBeforeFallback int

	// promotion
	levelProbeInterval time.Duration
	lastProbeAt        atomic.Int64

	stopCh   chan struct{}
	notifyCh chan struct{}
}

func (i *Ingestor[T]) Start() {
	i.stopCh = make(chan struct{})
	i.notifyCh = make(chan struct{}, 1)

	now := time.Now()
	i.lastFlush.Store(now.UnixNano())
	i.lastProbeAt.Store(now.UnixNano())

	go i.loop()
}

func (i *Ingestor[T]) Stop() {
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

func (i *Ingestor[T]) process() {
	now := time.Now()

	if i.rowsBuf.Len() == 0 {
		return
	}

	if i.rowsBuf.Len() >= i.maxFlushRows ||
		now.Sub(time.Unix(0, i.lastFlush.Load())) >= i.flushDelay {

		i.flushWithPolicy(now)
	}
}

func (i *Ingestor[T]) flushWithPolicy(now time.Time) {
	rows := i.rowsBuf.PeekBatch(i.maxFlushRows)
	if len(rows) == 0 {
		return
	}

	i.maybePromote(now)

	for lvl := i.currentLevel; lvl < len(i.levels); lvl++ {
		if lvl == i.currentLevel && now.UnixNano() < i.nextRetryAt.Load() {
			return
		}

		err := i.levels[lvl].Write(rows)
		if err == nil {
			i.currentLevel = lvl
			i.rowsBuf.CommitAndRelease(len(rows))
			i.lastFlush.Store(now.UnixNano())
			i.nextRetryAt.Store(0)
			return
		}

		i.logger.Warn(errors.Wrapf(err, "ingestion level %s failed", i.levels[lvl].Name()))

		// manage retry policy only on current level
		if lvl == i.currentLevel {
			i.nextRetryAt.Store(now.Add(i.retryDelay).UnixNano())

			if i.rowsBuf.Len() >= i.maxBufferBeforeFallback &&
				lvl+1 < len(i.levels) {

				i.logger.Warn(errors.WrapMsg("falling back to next ingestion level"))
				i.currentLevel++
				return
			}

			return
		}
	}

	i.logger.Error(errors.WrapMsg("all ingestion levels failed"))
}

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
