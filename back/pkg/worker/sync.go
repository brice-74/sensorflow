package worker

type PoolSync[Res any] interface {
	Submit(fn func() (Res, error)) (Res, error)
}

type ResultPoolSync[Res any] struct {
	pool PoolCore
}

var _ PoolSync[any] = (*ResultPoolSync[any])(nil)

func NewResultPoolSync[Res any](pool PoolCore) *ResultPoolSync[Res] {
	return &ResultPoolSync[Res]{pool: pool}
}

func (r *ResultPoolSync[Res]) Submit(fn func() (Res, error)) (Res, error) {
	ch := make(chan struct{}, 1)
	var res Res
	var err error

	sub := func() {
		res, err = fn()
		select {
		case ch <- struct{}{}:
		default:
		}
	}

	if submitErr := r.pool.Submit(VoidTask(sub)); submitErr != nil {
		return res, submitErr
	}

	<-ch
	return res, err
}
