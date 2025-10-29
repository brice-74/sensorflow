package redis

import (
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/redis/go-redis/v9"
)

type StringCmd[T any] struct {
	Cmd      *redis.StringCmd
	markDown func()
}

func (cmd StringCmd[T]) Result() (*T, error) {
	str, err := cmd.Cmd.Result()
	if err = HandleError(err, cmd.markDown); err != nil {
		return nil, err
	}
	return gobutil.Decode[T]([]byte(str))
}

type SliceCmd[T any] struct {
	Cmd      *redis.SliceCmd
	markDown func()
}

func (cmd SliceCmd[T]) Result() ([]*T, error) {
	slice, err := cmd.Cmd.Result()
	if err = HandleError(err, cmd.markDown); err != nil {
		return nil, err
	}
	return gobutil.DecodeManyAnyStr[T](slice)
}
