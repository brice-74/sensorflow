package orchestrators

import (
	"sync/atomic"
	"time"

	"github.com/brice-74/sensorflow/internal/adapters/clickhouse"
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
	repo     clickhouse.Repo[T]
	taskPool worker.ResultTaskSession[any]
	rowsBuf  xsync.BatchSlice[T]
	backup   BackupIngestor[T]

	maxFlushRows          int
	maxFlushWait          time.Duration
	maxRotationRows       int
	maxRotationWait       time.Duration
	lastFlush             time.Time
	lastRotation          time.Time
	rowsSinceLastFlush    int
	RowsSinceLastRotation int

	state atomic.Uint32 // IngestorState
}

func (ingestor *Ingestor[T]) Start() {
}

func (ingestor *Ingestor[T]) Submit(rows ...T) {
}
