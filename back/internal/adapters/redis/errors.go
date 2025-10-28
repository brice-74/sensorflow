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
		return errors.NewWrappedErr(errors.CodeNotFound, err, nil)
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, redis.ErrPoolTimeout):
		return errors.NewWrappedErr(errors.CodeTimeout, err, nil)
	case errors.Is(err, context.Canceled):
		return errors.NewWrappedErr(errors.CodeCanceled, err, nil)
	}

	return err
}
