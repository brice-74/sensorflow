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
