package redis

import (
	"github.com/brice-74/sensorflow/pkg/gobutil"
	"github.com/redis/go-redis/v9"
)

type StringCmd[T any] struct {
	Cmd *redis.StringCmd
}

func (cmd StringCmd[T]) Result() (*T, error) {
	str, err := cmd.Cmd.Result()
	if err != nil {
		return nil, err
	}
	return gobutil.Decode[T]([]byte(str))
}

type SliceCmd[T any] struct {
	Cmd *redis.SliceCmd
}

func (cmd SliceCmd[T]) Result() ([]*T, error) {
	slice, err := cmd.Cmd.Result()
	if err != nil {
		return nil, err
	}
	return gobutil.DecodeManyAnyStr[T](slice)
}
