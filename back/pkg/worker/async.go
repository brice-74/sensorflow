package worker

type PoolAsync[Res any] interface {
	Submit(fn func() Res) error
	Results() <-chan Res
	Close()
}

type ResultPoolAsync[Res any] struct {
	pool     PoolCore
	resultCh chan Res
}

func NewResultPoolAsync[Res any](pool PoolCore, size int) *ResultPoolAsync[Res] {
	return &ResultPoolAsync[Res]{
		pool:     pool,
		resultCh: make(chan Res, size),
	}
}

func (m *ResultPoolAsync[Res]) Submit(fn func() Res) error {
	return m.pool.Submit(&ResultTask[Res]{fn: fn, resultCh: m.resultCh})
}

func (m *ResultPoolAsync[Res]) Results() <-chan Res {
	return m.resultCh
}

func (m *ResultPoolAsync[Res]) Close() {
	close(m.resultCh)
}
