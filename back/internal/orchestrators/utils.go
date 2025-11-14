package orchestrators

import "github.com/brice-74/sensorflow/pkg/errors"

type ResultRedisCmd[T any] interface {
	Result() (T, error)
}

func tryGetRedisCmd[T any](
	cmd ResultRedisCmd[T],
	errs *[]error,
) (T, bool) {
	return tryGet(cmd.Result, errs)
}

func tryGet[T any](
	do func() (T, error),
	errs *[]error,
) (T, bool) {
	v, err := do()
	if err == nil {
		return v, true
	}

	if errors.Is(err, errors.ErrNotFound) {
		return *new(T), true
	}

	*errs = append(*errs, err)
	return *new(T), false
}
