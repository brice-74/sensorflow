package worker

// Task represents a unit of work to be executed by the pool.
type Task interface {
	Do()
}

type VoidTask func()

func (t VoidTask) Do() { t() }

type ResultTask[Res any] struct {
	fn       func() Res
	resultCh chan<- Res
}

func (t *ResultTask[Res]) Do() {
	t.resultCh <- t.fn()
}

type ResultTaskSession[Res any] struct {
	pool     Pool
	resultCh chan Res
}

func NewResultTaskSession[Res any](pool Pool, size int) *ResultTaskSession[Res] {
	return &ResultTaskSession[Res]{
		pool:     pool,
		resultCh: make(chan Res, size),
	}
}

func (m *ResultTaskSession[Res]) Submit(fn func() Res) error {
	return m.pool.Submit(&ResultTask[Res]{fn: fn, resultCh: m.resultCh})
}

func (m *ResultTaskSession[Res]) Results() <-chan Res {
	return m.resultCh
}

func (m *ResultTaskSession[Res]) Close() {
	close(m.resultCh)
}
