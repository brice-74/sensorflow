package redis

import (
	"context"

	"github.com/brice-74/sensorflow/pkg/errors"
	"github.com/redis/go-redis/v9"
)

func HandleError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, redis.Nil):
		return errors.NewError(errors.ErrNotFound, err, nil)
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, redis.ErrPoolTimeout):
		return errors.NewError(errors.ErrTimeout, err, nil)
	case errors.Is(err, context.Canceled):
		return errors.NewError(errors.ErrCanceled, err, nil)
	}

	return err
}
