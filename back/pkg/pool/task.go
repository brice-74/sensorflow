package pool

// Task represents a unit of work to be executed by the pool.
type Task interface {
	Do()
}

type VoidTask func()

func (t VoidTask) Do() { t() }

type ResultTask[T any] struct {
	fn       func() (T, error)
	resultCh chan<- TaskResult[T]
}

func (t *ResultTask[T]) Do() {
	res, err := t.fn()
	t.resultCh <- TaskResult[T]{Value: res, Err: err}
}

type (
	TaskResult[T any] struct {
		Value T
		Err   error
	}

	ResultTaskSession[T any] struct {
		pool     WorkerPool
		resultCh chan TaskResult[T]
	}
)

func NewResultTaskSession[T any](pool WorkerPool, size int) *ResultTaskSession[T] {
	return &ResultTaskSession[T]{
		pool:     pool,
		resultCh: make(chan TaskResult[T], size),
	}
}

func (m *ResultTaskSession[T]) Submit(fn func() (T, error)) error {
	return m.pool.Submit(&ResultTask[T]{fn: fn, resultCh: m.resultCh})
}

func (m *ResultTaskSession[T]) Results() <-chan TaskResult[T] {
	return m.resultCh
}

func (m *ResultTaskSession[T]) Close() {
	close(m.resultCh)
}
