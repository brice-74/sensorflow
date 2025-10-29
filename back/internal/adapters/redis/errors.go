package redis

import (
	"context"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/redis/go-redis/v9"
)

// HandleError processes a Redis error and marks the client down if needed.
func HandleError(err error, markDown func()) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, redis.Nil):
		return errors.NewWrappedErr(errors.CodeNotFound, err, nil)

	case errors.Is(err, redis.ErrPoolTimeout),
		errors.Is(err, redis.ErrClosed):
		markDown()
		return errors.NewWrappedErr(errors.CodeUnavailable, err, nil)

	case errors.Is(err, context.Canceled):
		return errors.NewWrappedErr(errors.CodeCanceled, err, nil)

	case errors.Is(err, context.DeadlineExceeded):
		return errors.NewWrappedErr(errors.CodeTimeout, err, nil)

	default:
		markDown()
		return err
	}
}
